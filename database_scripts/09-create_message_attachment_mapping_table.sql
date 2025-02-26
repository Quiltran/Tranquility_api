CREATE TABLE attachment_mapping (
    post_id INTEGER REFERENCES message(id) ON DELETE CASCADE,
    attachment_id INTEGER REFERENCES attachment(id) ON DELETE CASCADE
);