CREATE TABLE channel (
    id SERIAL PRIMARY KEY,
    name text,
    message_count integer DEFAULT 0,
    guild_id integer REFERENCES guild(id),
    created_date TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc'),
    updated_date TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc')
)