CREATE TABLE IF NOT EXISTS links (
    id           SERIAL PRIMARY KEY
    , code         VARCHAR(10)        NOT NULL UNIQUE
    , original_url TEXT               NOT NULL UNIQUE
    , created_at   TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);