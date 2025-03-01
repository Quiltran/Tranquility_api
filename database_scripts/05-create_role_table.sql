CREATE TABLE role (
    id SERIAL PRIMARY KEY,
    name TEXT,
    guild_id INTEGER REFERENCES guild(id) ON DELETE CASCADE,
    created_date TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc'),
    updated_date TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc')
);