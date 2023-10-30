-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TYPE action AS ENUM ('ACCRUAL', 'WITHDRAWAL');
CREATE TABLE IF NOT EXISTS balances (
    id serial PRIMARY KEY,
    action action NOT NULL,
    amount decimal NOT NULL,
    user_id integer REFERENCES users (id),
    order_number text,
    created_at timestamp DEFAULT NOW()
);

-- создание индексов
CREATE INDEX IF NOT EXISTS balance_create_idx ON balances (created_at);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP INDEX balance_create_idx;
DROP TABLE balances;
DROP TYPE action;