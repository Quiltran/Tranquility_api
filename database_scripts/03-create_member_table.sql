CREATE TABLE member (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES auth(id),
    guild_id INTEGER REFERENCES guild(id),
    user_who_added INTEGER REFERENCES auth(id),
    created_date TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc'),
    updated_date TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc'),
    UNIQUE (user_id, guild_id)
);