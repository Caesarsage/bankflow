-- Create separate databases for each service

-- Identity Service Database
CREATE DATABASE identity_db;

\c identity_db;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    refresh_token VARCHAR(500) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);

CREATE INDEX idx_sessions_user_id ON sessions (user_id);

CREATE INDEX idx_sessions_token ON sessions (refresh_token);

-- Customer Service Database
\c postgres;

CREATE DATABASE customer_db;

\c customer_db;

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id UUID NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    ssn_encrypted VARCHAR(255),
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(50),
    zip_code VARCHAR(20),
    country VARCHAR(100) DEFAULT 'USA',
    kyc_status VARCHAR(50) DEFAULT 'PENDING',
    kyc_verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE kyc_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    customer_id UUID NOT NULL REFERENCES customers (id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL,
    document_number VARCHAR(100),
    document_url VARCHAR(500) NOT NULL,
    status VARCHAR(50) DEFAULT 'PENDING',
    verified_by UUID,
    rejection_reason TEXT,
    uploaded_at TIMESTAMP DEFAULT NOW(),
    verified_at TIMESTAMP
);

CREATE INDEX idx_customers_user_id ON customers (user_id);

CREATE INDEX idx_kyc_customer_id ON kyc_documents (customer_id);

-- Account Service Database
\c postgres;

CREATE DATABASE account_db;

\c account_db;

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    account_number VARCHAR(20) UNIQUE NOT NULL,
    customer_id UUID NOT NULL,
    account_type VARCHAR(50) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    balance DECIMAL(15, 2) DEFAULT 0.00,
    available_balance DECIMAL(15, 2) DEFAULT 0.00,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    interest_rate DECIMAL(5, 2) DEFAULT 0.00,
    opened_at TIMESTAMP DEFAULT NOW(),
    closed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE account_holds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    account_id UUID NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    amount DECIMAL(15, 2) NOT NULL,
    reason VARCHAR(255) NOT NULL,
    transaction_ref VARCHAR(100),
    expires_at TIMESTAMP,
    released_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_accounts_customer ON accounts (customer_id);

CREATE INDEX idx_accounts_number ON accounts (account_number);

CREATE INDEX idx_accounts_status ON accounts (status);

CREATE INDEX idx_holds_account ON account_holds (account_id);

-- Transaction Service Database
\c postgres;

CREATE DATABASE transaction_db;

\c transaction_db;

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    transaction_ref VARCHAR(50) UNIQUE NOT NULL,
    from_account_id UUID,
    to_account_id UUID,
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    transaction_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'PENDING',
    description TEXT,
    metadata JSONB,
    fraud_score DECIMAL(5, 2),
    fraud_status VARCHAR(50),
    initiated_by UUID,
    processed_at TIMESTAMP,
    reversed_at TIMESTAMP,
    reversal_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE transaction_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    transaction_id UUID NOT NULL REFERENCES transactions (id) ON DELETE CASCADE,
    old_status VARCHAR(50),
    new_status VARCHAR(50),
    changed_by UUID,
    reason TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_transactions_from_account ON transactions (from_account_id);

CREATE INDEX idx_transactions_to_account ON transactions (to_account_id);

CREATE INDEX idx_transactions_status ON transactions (status);

CREATE INDEX idx_transactions_ref ON transactions (transaction_ref);

CREATE INDEX idx_transactions_created ON transactions (created_at);

-- Fraud Service Database
\c postgres;

CREATE DATABASE fraud_db;

\c fraud_db;

CREATE TABLE fraud_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    rule_name VARCHAR(100) NOT NULL,
    rule_type VARCHAR(50) NOT NULL,
    conditions JSONB NOT NULL,
    action VARCHAR(50) NOT NULL,
    score_impact INT DEFAULT 0,
    priority INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE fraud_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    transaction_id UUID,
    account_id UUID,
    customer_id UUID,
    fraud_score DECIMAL(5, 2) NOT NULL,
    risk_level VARCHAR(50) NOT NULL,
    flags JSONB,
    status VARCHAR(50) DEFAULT 'OPEN',
    assigned_to UUID,
    resolution TEXT,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE blacklist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    entity_type VARCHAR(50) NOT NULL,
    entity_value VARCHAR(255) NOT NULL,
    reason TEXT,
    added_by UUID,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_fraud_cases_transaction ON fraud_cases (transaction_id);

CREATE INDEX idx_fraud_cases_status ON fraud_cases (status);

CREATE INDEX idx_blacklist_entity ON blacklist (entity_type, entity_value);

-- Grant permissions
\c identity_db;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO bankflow;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO bankflow;

\c customer_db;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO bankflow;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO bankflow;

\c account_db;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO bankflow;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO bankflow;

\c transaction_db;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO bankflow;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO bankflow;

\c fraud_db;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO bankflow;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO bankflow;
