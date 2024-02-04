-- +migrate up
CREATE TABLE products (
    id INTEGER PRIMARY KEY,
    name VARCHAR(50),
    price DECIMAL(10,2),
    description TEXT
);

-- +migrate statement_begin
CREATE OR REPLACE FUNCTION add_product(name VARCHAR(50), price DECIMAL(10,2), description TEXT, category_id INTEGER)
RETURNS VOID AS $$
DECLARE
product_id INTEGER;
BEGIN
    INSERT INTO products (name, price, description) VALUES (name, price, description) RETURNING id INTO product_id;
    INSERT INTO product_categories (product_id, category_id) VALUES (product_id, category_id);
END;
$$ LANGUAGE plpgsql;
-- +migrate statement_end

-- +migrate down
DROP FUNCTION add_product(VARCHAR(50), DECIMAL(10,2), TEXT, INTEGER);
DROP TABLE products;