-- Helper function to convert snake_case to camelCase
CREATE OR REPLACE FUNCTION snake_to_camel(key text)
    RETURNS text
    LANGUAGE sql
    IMMUTABLE
AS $$
    WITH pascal AS (
        SELECT regexp_replace(initcap(replace(key, '_', ' ')), '\s', '', 'g') AS val
    )
    SELECT lower(left(val, 1)) || substring(val, 2) FROM pascal;
$$;

-- Convert a row to JSONB with camelCase keys
CREATE OR REPLACE FUNCTION camel(input_row anyelement)
    RETURNS jsonb
    LANGUAGE sql
AS $$
    SELECT COALESCE(
        jsonb_object_agg(snake_to_camel(key), value),
        '{}'::jsonb
    )
    FROM jsonb_each(to_jsonb(input_row));
$$;

-- create updated_at trigger function
CREATE OR REPLACE FUNCTION trigger_update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
