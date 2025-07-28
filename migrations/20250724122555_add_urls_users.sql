-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls
ADD COLUMN user_id INTEGER NOT NULL;

ALTER TABLE urls
ADD CONSTRAINT fk_user
FOREIGN KEY (user_id) REFERENCES users(id)
ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls
DROP CONSTRAINT fk_user;

ALTER TABLE urls
DROP COLUMN user_id;
-- +goose StatementEnd
