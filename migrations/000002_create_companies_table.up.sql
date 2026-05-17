CREATE TYPE company_type AS ENUM (
    'Corporations',
    'NonProfit',
    'Cooperative',
    'Sole Proprietorship'
);

CREATE TABLE IF NOT EXISTS companies (
    id                   UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                 VARCHAR(15)  NOT NULL,
    description          VARCHAR(3000) NOT NULL DEFAULT '',
    amount_of_employees  INTEGER      NOT NULL CHECK (amount_of_employees >= 0),
    registered           BOOLEAN      NOT NULL DEFAULT FALSE,
    type                 company_type NOT NULL,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_companies_name UNIQUE (name)
);

CREATE INDEX idx_companies_name ON companies (name);
CREATE INDEX idx_companies_type ON companies (type);
