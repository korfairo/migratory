-- +migrate up
SELECT count(1) FROM orders;

-- +migrate down
-- +migrate statement_begin
SELECT count(1) FROM orders;