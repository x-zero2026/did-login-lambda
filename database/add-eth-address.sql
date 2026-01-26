-- Migration: Add Ethereum address field to users table
-- Date: 2026-01-26
-- Description: Add eth_address field to support XZ Wallet

-- Step 1: Add eth_address column
ALTER TABLE users ADD COLUMN IF NOT EXISTS eth_address VARCHAR(42);

-- Step 2: Create index for eth_address
CREATE INDEX IF NOT EXISTS idx_users_eth_address ON users(eth_address);

-- Step 3: Add comment
COMMENT ON COLUMN users.eth_address IS 'Ethereum wallet address (42 chars: 0x + 40 hex) derived from BIP44 path m/44''/60''/0''/0/0';

-- Step 4: Verify the change
DO $$
DECLARE
    column_exists BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'eth_address'
    ) INTO column_exists;
    
    IF column_exists THEN
        RAISE NOTICE 'Migration successful! eth_address column added.';
    ELSE
        RAISE EXCEPTION 'Migration failed! eth_address column not found.';
    END IF;
END $$;

-- Migration complete
SELECT 'Migration completed successfully. Users table now has eth_address column.' AS status;
