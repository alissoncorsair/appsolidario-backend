CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(255) NOT NULL,
    payer_id INT REFERENCES "users"(id),
    payee_id INT REFERENCES "users"(id),
    amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);