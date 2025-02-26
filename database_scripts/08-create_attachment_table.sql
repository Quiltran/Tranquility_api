CREATE TABLE attachment (
    id SERIAL PRIMARY KEY,
    file_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size DECIMAL(10, 4),
    mime_type TEXT NOT NULL,
    user_uploaded INTEGER NOT NULL,
    created_date TIMESTAMPTZ DEFAULT (NOW() at TIME ZONE 'utc')
);