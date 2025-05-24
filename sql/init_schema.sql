-- Create token changes trigger
CREATE OR REPLACE FUNCTION tibber.notify_token_changes()
RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify(
        'token_changes',
        json_build_object(
            'action', TG_OP,
            'token_id', NEW.id,
            'active', NEW.active
        )::text
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
DROP TRIGGER IF EXISTS token_changes_trigger ON tibber.tibber_tokens;
CREATE TRIGGER token_changes_trigger
    AFTER INSERT OR UPDATE ON tibber.tibber_tokens
    FOR EACH ROW
    EXECUTE FUNCTION tibber.notify_token_changes(); 