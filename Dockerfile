FROM golang:1.13-alpine as build
RUN apk update && apk add --no-cache git make
COPY . /workspace
WORKDIR /workspace
ARG _TAG
ENV VERSION ${_TAG}
RUN mkdir -p bin
RUN make build

FROM alpine:latest
COPY --from=build /workspace/bin/roster /bin/roster
EXPOSE 8080
