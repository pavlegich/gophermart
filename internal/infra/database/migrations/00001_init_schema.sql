-- +goose Up
-- создание таблицы users
CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    login text UNIQUE NOT NULL,
    password bytea NOT NULL
);

-- создание таблицы orders
CREATE TYPE status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE IF NOT EXISTS orders (
    id serial PRIMARY KEY,
    number text UNIQUE NOT NULL,
    user_id integer REFERENCES users (id),
    status status DEFAULT 'NEW',
    accrual decimal DEFAULT 0,
    created_at timestamp DEFAULT NOW()
);

-- создание таблицы balances
CREATE TYPE action AS ENUM ('ACCRUAL', 'WITHDRAWAL');
CREATE TABLE IF NOT EXISTS balances (
    id serial PRIMARY KEY,
    action action,
    amount decimal,
    user_id integer REFERENCES users (id),
    order_id integer NOT NULL REFERENCES orders (id),
    created_at timestamp DEFAULT NOW()
);

-- создание индексов для таблиц orders и balances по полю created_at
CREATE INDEX IF NOT EXISTS order_create_idx ON orders (created_at);
CREATE INDEX IF NOT EXISTS balance_create_idx ON balances (created_at);

-- +goose Down
DROP INDEX order_create_idx;
DROP INDEX balance_create_idx;
DROP TABLE balances;
DROP TYPE action;
DROP TABLE orders;
DROP TYPE status;
DROP TABLE users;