CREATE TABLE IF NOT EXISTS users 
(
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    organization_id INT NULL,
    phone VARCHAR (255) NOT NULL,
    password VARCHAR(510) NOT NULL,
    avatar_url VARCHAR(255) NULL,
    balance INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP 
);

CREATE INDEX IF NOT EXISTS idx_phone ON users (phone);
