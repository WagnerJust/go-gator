-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE feeds (
    id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name text NOT NULL,
    url text UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
    PRIMARY KEY(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE feeds;
-- +goose StatementEnd
