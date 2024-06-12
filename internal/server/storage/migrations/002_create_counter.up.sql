CREATE TABLE IF NOT EXISTS counter
(
    id    SERIAL PRIMARY KEY,
    name  VARCHAR(255) UNIQUE NOT NULL,
    value BIGINT              NOT NULL
);