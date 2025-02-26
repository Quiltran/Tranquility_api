CREATE TABLE role_intent (
    id SERIAL PRIMARY KEY,
    role_id INTEGER REFERENCES role(id),
    value INTEGER,
    created_date TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc'),
    updated_date TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc')
);