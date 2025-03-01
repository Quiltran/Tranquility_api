CREATE TABLE guild (
    id SERIAL PRIMARY KEY,
    name TEXT,
    description TEXT,
    owner_id integer REFERENCES auth(id) ON DELETE CASCADE,
    created_date TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc'),
    updated_date TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc')
)