-- +migrate up
SELECT count(1) FROM orders;

-- +migrate
SELECT count(1) FROM orders;