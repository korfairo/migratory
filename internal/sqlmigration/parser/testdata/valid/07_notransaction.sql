-- +migrate up no_transaction
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    hire_date DATE NOT NULL DEFAULT CURRENT_DATE,
    salary DECIMAL(10,2) NOT NULL
);

CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    manager_id INTEGER REFERENCES employees(id) ON DELETE SET NULL
);

ALTER TABLE employees ADD COLUMN created_at TIMESTAMP DEFAULT NOW();

CREATE INDEX idx_departments_manager_id ON departments(manager_id);

-- +migrate down no_transaction
DROP TABLE departments;
DROP TABLE employees;