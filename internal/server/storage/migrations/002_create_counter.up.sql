CREATE TABLE IF NOT EXISTS counter
(
    id    SERIAL PRIMARY KEY,
    name  VARCHAR(255) UNIQUE NOT NULL,
    value INT                 NOT NULL
);