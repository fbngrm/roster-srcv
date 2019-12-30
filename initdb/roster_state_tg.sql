CREATE OR REPLACE FUNCTION active_players_per_roster() RETURNS TRIGGER AS $pl$
DECLARE
    n integer;
BEGIN
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        SELECT INTO n count(id) FROM players WHERE roster_id = NEW.roster_id AND status = 'active';
        IF n <> 5 THEN
            RAISE EXCEPTION 'During % of players: roster id=% must have exactly 5 active players, not %',tg_op,NEW.roster_id,n;
        END IF;
    END IF;

    IF TG_OP = 'UPDATE' OR TG_OP = 'DELETE' THEN
        SELECT INTO n count(id) FROM players WHERE roster_id = OLD.roster_id AND status = 'active';
        IF n <> 5 THEN
            RAISE EXCEPTION 'During % of players: roster id=% must have exactly 5 active players, not %',tg_op,NEW.roster_id,n;
        END IF;
    END IF;

    RETURN NULL;
END;
$pl$ LANGUAGE 'plpgsql';

CREATE CONSTRAINT TRIGGER active_players_per_roster
AFTER INSERT OR UPDATE OR DELETE ON players
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE PROCEDURE active_players_per_roster();

CREATE OR REPLACE FUNCTION roster_constrain_active_players() RETURNS trigger AS $ro$
DECLARE
    n integer;
BEGIN
    IF TG_OP = 'INSERT' THEN
        SELECT INTO n count(id) FROM players WHERE roster_id = NEW.id AND status = 'active';
        IF n <> 5 THEN
            RAISE EXCEPTION 'During INSERT of roster id=%: Must have 5 active players, found %',NEW.id,n;
        END IF;
    END IF;
    -- No need for an UPDATE or DELETE check, as regular referential integrity constraints
    -- and the trigger on `players' will do the job.
    RETURN NULL;
END;
$ro$ LANGUAGE 'plpgsql';


CREATE CONSTRAINT TRIGGER roster_limit_players_tg
AFTER INSERT ON rosters
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE PROCEDURE roster_constrain_active_players();
