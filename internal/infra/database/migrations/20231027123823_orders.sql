-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TYPE status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE IF NOT EXISTS orders (
    id serial PRIMARY KEY,
    number text UNIQUE NOT NULL,
    user_id integer REFERENCES users (id),
    status status DEFAULT 'NEW',
    accrual decimal DEFAULT 0,
    created_at timestamp DEFAULT NOW()
);

-- создание индексов
CREATE INDEX IF NOT EXISTS order_user_id_idx ON orders (user_id);
CREATE INDEX IF NOT EXISTS order_create_idx ON orders (created_at);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP INDEX order_create_idx;
DROP INDEX order_user_id_idx;
DROP TABLE orders;
DROP TYPE status;