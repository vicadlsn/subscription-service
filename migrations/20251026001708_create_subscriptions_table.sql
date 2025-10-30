-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    service_name VARCHAR(64) NOT NULL,
    price INT NOT NULL CHECK (price > 0),
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_subscriptions_user_id
    ON subscriptions(user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_subscriptions_service_name
    ON subscriptions(service_name)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_subscriptions_start_date
    ON subscriptions(start_date)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_subscriptions_user_service_start
    ON subscriptions(user_id, service_name, start_date)
    WHERE deleted_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE subscriptions;
-- +goose StatementEnd
