-- +goose Up
CREATE TABLE IF NOT EXISTS user (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    login      TEXT NOT NULL UNIQUE,
    email      TEXT NOT NULL UNIQUE,
    pass       TEXT NOT NULL, 
    status     TEXT NOT NULL DEFAULT 'INVALID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    CHECK (status = 'INVALID' OR status = 'VALID' OR status = 'BANNED')
);

CREATE TABLE IF NOT EXISTS capture (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER,
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    type        TEXT NOT NULL,
    has_cab     BOOL,
    am2_hash    TEXT NOT NULL,
    data_hash   TEXT NOT NULL UNIQUE,
    data        BLOB NOT NULL,
    downloads   INTEGER NOT NULL DEFAULT 0,
    demo_link   TEXT,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME,
    FOREIGN KEY (user_id)
        REFERENCES user(id)
            ON DELETE SET NULL,
    CHECK (type = 'CLEAN' OR type = 'CRUNCH' OR type = 'HI-GAIN')
);

CREATE TABLE IF NOT EXISTS user_favorite (    
    user_id    INTEGER NOT NULL,
    capture_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, capture_id),
    FOREIGN KEY (user_id)
        REFERENCES user(id)
            ON DELETE CASCADE,
    FOREIGN KEY (capture_id)
        REFERENCES capture(id)
            ON DELETE CASCADE 
);

CREATE TABLE IF NOT EXISTS review (
    id         INTEGER NOT NULL PRIMARY KEY,
    user_id    INTEGER,
    capture_id INTEGER NOT NULL,
    rate       INTEGER NOT NULL,
    comment    TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    FOREIGN KEY (user_id)
        REFERENCES user(id)
            ON DELETE SET NULL,
    FOREIGN KEY (capture_id)
        REFERENCES capture(id)
            ON DELETE CASCADE      
);

CREATE TABLE IF NOT EXISTS protected_am2 (
    am2_hash   TEXT NOT NULL PRIMARY KEY,
    ref        TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (length(am2_hash) = 64)
);

-- +goose Down
DROP TABLE IF EXISTS protected_am2;
DROP TABLE IF EXISTS review;
DROP TABLE IF EXISTS user_favorite;
DROP TABLE IF EXISTS capture;
DROP TABLE IF EXISTS user;
