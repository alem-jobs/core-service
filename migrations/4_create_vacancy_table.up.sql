CREATE TABLE vacancies (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    salary_from DECIMAL(10,2) NULL,
    salary_to DECIMAL(10,2) NULL,
    salary_exact DECIMAL(10,2) NULL,
    salary_type VARCHAR(255) NOT NULL,
    salary_currency VARCHAR(255) NOT NULL,
    organization_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE vacancy_details (
    id SERIAL PRIMARY KEY,
    group_name VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    value VARCHAR(255) NOT NULL,
    icon_url VARCHAR(255) NULL,
    vacancy_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_vacancy FOREIGN KEY (vacancy_id) REFERENCES vacancies(id) ON DELETE CASCADE
);

