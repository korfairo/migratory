-- +migrate up
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    description TEXT
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_name VARCHAR(50) NOT NULL,
    order_date DATE NOT NULL DEFAULT CURRENT_DATE,
    total_price DECIMAL(10,2) NOT NULL
);

ALTER TABLE products ADD COLUMN created_at TIMESTAMP DEFAULT NOW();

-- +migrate down
DROP TABLE orders;
DROP TABLE products;