-- +goose Up
ALTER TABLE user
    ADD COLUMN picture TEXT;

-- +goose Down
ALTER TABLE user
    DROP COLUMN picture;