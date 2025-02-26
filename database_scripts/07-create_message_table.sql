CREATE TABLE message (
    id SERIAL PRIMARY KEY,
    channel_id INTEGER REFERENCES channel(id),
    author_id INTEGER REFERENCES auth(id),
    content TEXT,
    created_date TIMESTAMPTZ DEFAULT (NOW() at TIME ZONE 'utc'),
    updated_date TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc')
);

CREATE INDEX idx_messages_create_date_desc ON message (created_date DESC);