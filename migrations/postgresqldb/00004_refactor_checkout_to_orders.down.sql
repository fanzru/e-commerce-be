-- Rollback checkout refactoring
BEGIN;

-- Drop the view
DROP VIEW IF EXISTS order_history;

-- Remove added columns
ALTER TABLE checkouts 
DROP COLUMN IF EXISTS user_id,
DROP COLUMN IF EXISTS payment_status,
DROP COLUMN IF EXISTS payment_method,
DROP COLUMN IF EXISTS payment_reference,
DROP COLUMN IF EXISTS notes,
DROP COLUMN IF EXISTS status,
DROP COLUMN IF EXISTS completed_at;

-- Drop indexes
DROP INDEX IF EXISTS idx_checkouts_user_id;
DROP INDEX IF EXISTS idx_checkouts_payment_status;
DROP INDEX IF EXISTS idx_checkouts_status;

COMMIT; 