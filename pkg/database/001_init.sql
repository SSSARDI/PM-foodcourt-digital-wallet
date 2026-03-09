DROP DATABASE IF EXISTS foodcourt;
CREATE DATABASE foodcourt;
USE foodcourt;
CREATE DATABASE IF NOT EXISTS foodcourt;
USE foodcourt;

-- USERS
CREATE TABLE users (
    id CHAR(36) PRIMARY KEY,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    password_hash TEXT,
    role ENUM('CUSTOMER','VENDOR','STAFF','ADMIN') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status ENUM('ACTIVE','SUSPENDED') DEFAULT 'ACTIVE'
);

-- WALLETS
CREATE TABLE wallets (
    id CHAR(36) PRIMARY KEY,
    public_id CHAR(36) UNIQUE NOT NULL,
    user_id CHAR(36) UNIQUE NOT NULL,
    balance DECIMAL(12,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- FOOD STALLS
CREATE TABLE food_stalls (
    id CHAR(36) PRIMARY KEY,
    stall_name VARCHAR(100) NOT NULL,
    owner_id CHAR(36) NOT NULL,
    category VARCHAR(100),
    location VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status ENUM('OPEN','CLOSED'),
    FOREIGN KEY (owner_id) REFERENCES users(id)
);

-- WALLET TRANSACTIONS
CREATE TABLE wallet_transactions (
    id CHAR(36) PRIMARY KEY,
    wallet_id CHAR(36) NOT NULL,
    type ENUM('TOPUP_QR','TOPUP_CASH','PAYMENT','REFUND') NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    reference_id CHAR(36),
    created_by CHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

-- SALE TRANSACTIONS
CREATE TABLE sale_transactions (
    id CHAR(36) PRIMARY KEY,
    stall_id CHAR(36) NOT NULL,
    customer_id CHAR(36) NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL,
    payment_transaction_id CHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (stall_id) REFERENCES food_stalls(id),
    FOREIGN KEY (customer_id) REFERENCES users(id),
    FOREIGN KEY (payment_transaction_id) REFERENCES wallet_transactions(id)
);

-- QR PAYMENTS
CREATE TABLE qr_payments (
    id CHAR(36) PRIMARY KEY,
    stall_id CHAR(36) NOT NULL,
    qr_token VARCHAR(255) UNIQUE,
    amount DECIMAL(12,2),
    expires_at TIMESTAMP,
    status ENUM('PENDING','PAID','EXPIRED'),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (stall_id) REFERENCES food_stalls(id)
);

-- PENDING CASH REFUNDS
CREATE TABLE pending_cash_refunds (
    id CHAR(36) PRIMARY KEY,
    wallet_id CHAR(36) NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    status ENUM('PENDING','APPROVED','REJECTED','EXPIRED'),
    initiated_by CHAR(36),
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id)
);

-- GP VAT
CREATE TABLE gp_vat (
    gp DECIMAL(12,2) NOT NULL,
    vat DECIMAL(12,2) NOT NULL,
    PRIMARY KEY (gp, vat)
);

-- PASSWORD RESET TOKENS
CREATE TABLE password_reset_tokens (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
