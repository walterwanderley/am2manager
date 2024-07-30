-- +goose Up
DROP TABLE user;

CREATE TABLE IF NOT EXISTS user (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,    
    email      TEXT NOT NULL UNIQUE,
    name       TEXT NOT NULL UNIQUE,
    status     TEXT NOT NULL DEFAULT 'VALID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    CHECK (status = 'INVALID' OR status = 'VALID')
);

-- +goose Down
