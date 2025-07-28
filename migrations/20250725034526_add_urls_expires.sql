-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls
ADD COLUMN expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '1 year');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls
DROP COLUMN expires_at;
-- +goose StatementEnd
