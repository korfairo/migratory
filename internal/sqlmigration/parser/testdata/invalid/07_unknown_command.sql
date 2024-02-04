-- +migrate apply
SELECT count(1) FROM orders;

-- +migrate rollback
SELECT count(1) FROM orders;