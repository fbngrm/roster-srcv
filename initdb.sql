CREATE TABLE rosters (
    id   BIGSERIAL PRIMARY KEY,
    name varchar(32) NOT NULL
);

CREATE TABLE players (
    id         BIGSERIAL PRIMARY KEY,
    roster_id  BIGINT REFERENCES rosters(id),
    first_name varchar(32) NOT NULL,
    last_name  varchar(32) NOT NULL,
    alias      varchar(32) NOT NULL,
    active     BOOLEAN
);

INSERT INTO rosters(id,name)
VALUES(382574876546039808,'foo');

INSERT INTO players(id,roster_id,first_name,last_name,alias,active) VALUES
(182919996442279937,382574876546039808,'Dominic','Dominic','DataSlayer9',true),
(337332768876789763,382574876546039808,'Jane','Beddingfield','__Jain',true),
(444322878230495243,382574876546039808,'Phillip','Aaronivic','phikic',true),
(602403447886839809,382574876546039808,'Ji','Bhok','TARG3T',true),
(622318474387128331,382574876546039808,'Damian','Grey','Klikx',true),
(184315303323238400,382574876546039808,'Oliver','Fieldbutter','Smaayo',false);
