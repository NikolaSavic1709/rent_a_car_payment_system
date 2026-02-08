-- Crypto Payment Microservice Database Schema

-- Create crypto_payments table
CREATE TABLE IF NOT EXISTS crypto_payments (
    id SERIAL PRIMARY KEY,
    payment_id UUID UNIQUE NOT NULL,
    transaction_id UUID NOT NULL,
    merchant_order_id UUID NOT NULL,
    merchant_id INTEGER NOT NULL,
    amount DECIMAL(18, 8) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    status INTEGER NOT NULL DEFAULT 0,
    destination_address VARCHAR(255) NOT NULL,
    source_address VARCHAR(255),
    tx_hash VARCHAR(255),
    block_height BIGINT,
    confirmations INTEGER DEFAULT 0,
    required_confirmations INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expiry_time TIMESTAMP NOT NULL,
    confirmed_at TIMESTAMP,
    is_testnet BOOLEAN DEFAULT true
);

-- Create indexes for crypto_payments
CREATE INDEX IF NOT EXISTS idx_crypto_payments_payment_id ON crypto_payments(payment_id);
CREATE INDEX IF NOT EXISTS idx_crypto_payments_transaction_id ON crypto_payments(transaction_id);
CREATE INDEX IF NOT EXISTS idx_crypto_payments_merchant_order_id ON crypto_payments(merchant_order_id);
CREATE INDEX IF NOT EXISTS idx_crypto_payments_status ON crypto_payments(status);
CREATE INDEX IF NOT EXISTS idx_crypto_payments_tx_hash ON crypto_payments(tx_hash);

-- Create merchant_wallets table
CREATE TABLE IF NOT EXISTS merchant_wallets (
    id SERIAL PRIMARY KEY,
    merchant_id INTEGER NOT NULL,
    currency VARCHAR(10) NOT NULL,
    wallet_address VARCHAR(255) NOT NULL UNIQUE,
    public_key TEXT NOT NULL,
    private_key TEXT,
    balance DECIMAL(18, 8) DEFAULT 0,
    is_testnet BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(merchant_id, currency)
);

-- Create indexes for merchant_wallets
CREATE INDEX IF NOT EXISTS idx_merchant_wallets_merchant_id ON merchant_wallets(merchant_id);
CREATE INDEX IF NOT EXISTS idx_merchant_wallets_currency ON merchant_wallets(currency);
CREATE INDEX IF NOT EXISTS idx_merchant_wallets_address ON merchant_wallets(wallet_address);

-- Create blockchain_transactions table (optional, for detailed tracking)
CREATE TABLE IF NOT EXISTS blockchain_transactions (
    id SERIAL PRIMARY KEY,
    tx_hash VARCHAR(255) UNIQUE NOT NULL,
    payment_id UUID NOT NULL,
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    amount DECIMAL(18, 8) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    block_height BIGINT,
    confirmations INTEGER DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    detected_at TIMESTAMP NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMP,
    FOREIGN KEY (payment_id) REFERENCES crypto_payments(payment_id)
);

-- Create indexes for blockchain_transactions
CREATE INDEX IF NOT EXISTS idx_blockchain_tx_hash ON blockchain_transactions(tx_hash);
CREATE INDEX IF NOT EXISTS idx_blockchain_payment_id ON blockchain_transactions(payment_id);
CREATE INDEX IF NOT EXISTS idx_blockchain_status ON blockchain_transactions(status);

-- Insert some test data for development
-- Test merchant wallets
INSERT INTO merchant_wallets (merchant_id, currency, wallet_address, public_key, balance, is_testnet, created_at, updated_at)
VALUES 
    (12345, 'BTC', 'tb1qw508d6qejxtdg4y5r3zarvary0c5xw7kxpjzsx', '02e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855', 0, true, NOW(), NOW()),
    (12345, 'ETH', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb', '04e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855', 0, true, NOW(), NOW()),
    (12345, 'USDT', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEc', '04e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b856', 0, true, NOW(), NOW())
ON CONFLICT (merchant_id, currency) DO NOTHING;

-- Add comments for documentation
COMMENT ON TABLE crypto_payments IS 'Stores all cryptocurrency payment transactions';
COMMENT ON TABLE merchant_wallets IS 'Stores merchant cryptocurrency wallet addresses';
COMMENT ON TABLE blockchain_transactions IS 'Detailed tracking of blockchain transactions';

COMMENT ON COLUMN crypto_payments.status IS '0=Pending, 1=Confirming, 2=Confirmed, 3=Expired, 4=Failed';
COMMENT ON COLUMN crypto_payments.amount IS 'Amount in cryptocurrency (8 decimal places)';
COMMENT ON COLUMN crypto_payments.required_confirmations IS 'Number of confirmations needed (3 for BTC, 12 for ETH)';

COMMENT ON COLUMN merchant_wallets.private_key IS 'Encrypted private key (should never be exposed)';
COMMENT ON COLUMN merchant_wallets.balance IS 'Current wallet balance (for tracking only)';
