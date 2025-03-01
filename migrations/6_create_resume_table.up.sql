CREATE TABLE IF NOT EXISTS resumes (
    id SERIAL PRIMARY KEY,
    user_id INT,
    category_id INT NOT NULL,
    description VARCHAR(10000) NULL,
    salary_from INT NULL,
    salary_to INT NULL,
    salary_period VARCHAR(255) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS resume_skills (
    id SERIAL PRIMARY KEY,
    resume_id INT,
    skill VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS resume_experiences (
    id SERIAL PRIMARY KEY,
    resume_id INT NOT NULL,
    oraganization_name VARCHAR(255) NOT NULL,
    category_id INT NOT NULL,
    description VARCHAR(1000) NOT NULL,
    start_month VARCHAR(255) NOT NULL,
    start_year VARCHAR(255) NOT NULL,
    end_month VARCHAR(255) NULL,
    end_year VARCHAR(255) NULL
);

