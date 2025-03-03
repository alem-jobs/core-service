CREATE TABLE organization_services (
       id SERIAL PRIMARY KEY,
       name VARCHAR(255),
       description VARCHAR(1000) NULL,
       organization_id INT NOT NULL,
       price_from NUMERIC(10, 2) NOT NULL,
       price_to NUMERIC(10, 2) NOT NULL,
       deadline VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS reviews (
       id SERIAL PRIMARY KEY,
       user_id INT NOT NULL,
       company_id INT NOT NULL,
       mark INT NOT NULL,
       content VARCHAR(1000) NULL,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
