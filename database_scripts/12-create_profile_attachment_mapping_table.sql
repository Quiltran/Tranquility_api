CREATE TABLE profile_mapping(
    user_id INTEGER UNIQUE REFERENCES auth(id) ON DELETE CASCADE,
    attachment_id INTEGER REFERENCES attachment(id) ON DELETE CASCADE
);