CREATE TABLE rosters (
    id   BIGSERIAL PRIMARY KEY,
    name varchar(32) UNIQUE NOT NULL
);

CREATE TABLE players (
    id         BIGSERIAL PRIMARY KEY,
    roster_id  BIGINT REFERENCES rosters(id) NOT NULL,
    first_name varchar(32) NOT NULL,
    last_name  varchar(32) NOT NULL,
    alias      varchar(32) NOT NULL,
    status     varchar(32) NOT NULL
);

INSERT INTO rosters(id,name) VALUES
(382574876546039808,'foo'),
(382574876546039807,'bar');

INSERT INTO players(id,roster_id,first_name,last_name,alias,status) VALUES
(182919996442279937,382574876546039808,'Dominic','Luklowski','DataSlayer9','active'),
(337332768876789763,382574876546039808,'Jane','Beddingfield','__Jain','active'),
(444322878230495243,382574876546039808,'Phillip','Aaronivic','phikic','active'),
(602403447886839809,382574876546039808,'Ji','Bhok','TARG3T','active'),
(622318474387128331,382574876546039808,'Damian','Grey','Klikx','active'),
(184315303323238400,382574876546039808,'Oliver','Fieldbutter','Smaayo','benched');
