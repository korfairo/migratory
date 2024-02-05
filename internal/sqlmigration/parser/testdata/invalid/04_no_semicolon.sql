-- +migrate down
SELECT count(1) FROM orders

-- +migrate up
SELECT count(1) FROM orders;