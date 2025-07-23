-- +goose Up
-- +goose StatementBegin
CREATE TABLE urls (
	id SERIAL PRIMARY KEY,
	long_url TEXT NOT NULL,
	short_url VARCHAR(255) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS urls;
-- +goose StatementEnd
