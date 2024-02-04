-- +migrate up
SELECT count(1) FROM orders;
-- +migrate statement_end

-- +migrate down
SELECT count(1) FROM orders;