CREATE TABLE guild (
    id SERIAL PRIMARY KEY,
    name TEXT,
    description TEXT,
    owner_id integer REFERENCES auth(id),
    created_date TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc'),
    updated_date TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc')
)