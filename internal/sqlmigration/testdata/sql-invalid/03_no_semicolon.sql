-- +migrate up
SELECT count(1) FROM orders
-- +migrate up

-- +migrate down
SELECT count(1) FROM orders
-- +migrate down