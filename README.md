# Roster change
This branch implements the requirements defined in the [task](https://github.com/fbngrm/roster-srcv/blob/master/Task_-_Roster_Change.pdf) description.

## Overview
This document is organized in three sections:

* Documentation - describes setup and usage
* Architecture - describes the design and architecture approach
* Optimizations and Bugs - describes approaches to enhance the code as well as known bugs

## Documentation

### Setup
This section assumes there is a go, docker, make and git installation available on the system.

To check your installation, run:

```bash
go version
docker version
make --version
git version
```

Fetch the repo from GitHub:

```bash
cd $GOPATH/src/fbngrm/roster-srcv
git clone git@github.com:fbngrm/roster-srcv.git
cd roster-srcv
```

### Dependency management
For handling dependencies, go modules are used.
This requires to have a go version > 1.11 installed and setting `GO111MODULE=1`.
If the go version is >= 1.13, modules are enabled by default.
There might be steps required to access private repositories.
If you have problems setting up or building the project which are related to modules, please consider reading up the [documentation](https://github.com/golang/go/wiki/Modules).
If this does not solve the issue problem open an issue here.

### Usage
A Makefile are provided which should be used to test, build and run the service.
The service is started in a docker container.
The configuration resides in the [docker-compose](https://github.com/fbngrm/roster-srcv/blob/master/docker-compose.yaml) file.
The Dockerfile used to build images is located in the project root.

### Build
Builds will be placed in the `/bin` directory. Binaries use the latest git commit hash or tag as a version.

### Run
The service is intended to be ran in a docker container.

```bash
make build
make run
```

### API
The provided API methods are HTTP/1.1 compliant according to RFC2616 and RFC5789.

PATCH endpoints expect request payloads to be formatted according to the `JSON Merge Patch`
definition of RFC7396.

#### Add a player
The application supports adding of new players.
The endpoint expects a POST requests with a JSON payload containing the player data.
If supplied, the player-id will be ignored.
Instead, it is generated by the datastore and returned with the complete player representation in JSON format on success.

`POST /players/add`

```bash
curl -X POST http://127.0.0.1:8080/players/add \
    -H "Content-Type: application/json" \
    -d '{"roster_id":382574876546039808,"first_name":"foo","last_name":"bar","alias":"foobar"}'
```

#### Add a player to the roster
To add a player to a roster, a PATCH request must be used since a partial update is performed to an existing resource.
The request payload needs to contain the new roster-id.
A JSON representation of the updated player is returned.
An error is returned it the roster does not exist.
When adding a player to a new roster, the player is benched by default to not corrupt the roster's state.

`PATCH /players/update`

```bash
curl -i -X PATCH http://127.0.0.1:8080/players/update \
    -H "Content-Type: application/json" \
    -d '{"player_id":444322878230495243,"roster_id":382574876546039808}'
```

#### Move a player to the active roster
When activating a player, another active player needs to be benched to keep the roster in a valid state.
The endpoint expects a PATCH request with a JSON payload containing two player ids of which one must be active and one must be benched.
Both players must be in the same roster.
If successful, a JSON representation of the updated players is returned.

`PATCH /players/change`

```bash
curl -i -X PATCH http://127.0.0.1:8080/players/change \
    -H "Content-Type: application/json" \
    -d '{"active":{"player_id":444322878230495243},"benched":{"player_id":184315303323238400}}'
```

#### Fetch the entire roster
A JSON representation or the entire roster can be retrieved via a GET request.
The roster is identified by the provided id in the URL path.

`GET /roster/:id`

```bash
curl -X GET http://127.0.0.1:8080/roster/382574876546039808
```

#### Fetch benched/active players
A JSON representation or the benched and active players of a roster can be retrieved
via a GET request.

`GET /roster/:id/:status`

```bash
curl -X GET http://127.0.0.1:8080/roster/382574876546039808/benched
```

### Tests
There are several targets available to run tests.

```bash
make test # runs tests
make test-cover # creates a coverage profile
make test-race # tests service for race conditions
```

### Lint
There is a lint target which runs [golangci-lint](https://github.com/golangci/golangci-lint) in a docker container.

```bash
make lint
```

### Code changes
After making changes to the code, it is required to rebuild the image(s):

```bash
docker-compose up --detach --build <service_name>
```

### Configuration
The service can be configured by parameters or environment variables.
For configuring the service via environment variables the docker-compose file should be used.
Alternatively, arguments can be supplied to the command directly.

#### Logging
The current setup uses a human friendly logging format.
The logger attaches the service name and build version to the log output.

## Architecture
I follow industry standards like the go [conventions](https://golang.org/doc/code.html), [proverbs](https://go-proverbs.github.io/) as well as the [12 Factor-App](https://12factor.net/) principles.

The interfaces are kept small to bigger the abstraction.
Variable names are short when they are used close to their declaration.
They are more meaningful if they are used outside the scope they were defined.
Errors are used as values.

Furthermore, I followed the dependency injection and fail early approach.
Components are provided all dependencies they need during instantiation.
The result is either a functioning instance or an error.
On application start-up, an error results in termination.
Runtime errors do not lead to a crash or panic.

Configuration is injected at start-up.

Termination signals lead to a graceful shutdown.
Meaning, all servers and handlers stop accepting new requests, process their current workload and shut down.
Though, there is a configurable shutdown timeout, which may prevent this.

### Tests
There are unit tests for core functionality, things expected to break and for edge/error cases.
In general I think testing on package boundaries as well as core functionality internally is a better approach than just aiming for a certain percentage of coverage.

### Testdata
There is a testdata directory which provides sample data used in tests.

### Dependencies
In general, the code is written in a way to use as few dependencies as possible and make use of the standard library whenever possible.
I try to avoid external test libraries, since they mostly do not provide significant advantages but may obfuscate clear readability.
This especially applies to BDD (Behavioral Driven Design/Development) test libraries, which often introduce needless indirection and conceptual overhead.

I used a muxer lib for convenience and validating routes and query parameters.
User supplied URL and query parameters must be treated as potentially malicious and should be sanitized using a production proven solution.

I also use a logger library that provides structured, leveled and sampled logging functionality, which I consider the minimum requirement for runtime monitoring.

### Database schema
The database is initialized by the scripts contained in the `initdb` directory.
See the [documentation](https://hub.docker.com/_/postgres) for more information.

Rosters and players are referenced by an `1:n` relationship.

The consistent state of a roster with no more and no less than 5 active players is ensured by a trigger. Meaning, a roster needed to be initialized in the above mentioned state.
Further, the provided API operations ensure that the state stays consistent.

### Optimizations
The error handling of errors returned by the datastore is not fine grained and
does not provide useful feedback to a user. This needed to be enhanced in a production setup.

To properly represent the status of a player in the database, we needed
to add another table that holds the valid states "active" and "benched", that get then
referenced by the players status column.

### Bugs
Adding an "active" player to the roster she currently is a member of, results in
the players status getting set to "benched".

