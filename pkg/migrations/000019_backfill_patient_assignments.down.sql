-- Down migration: Remove backfilled assignments
-- Note: This will remove all assignments, not just the backfilled ones
-- Use with caution in production

-- This migration cannot be safely reversed without losing data
-- We'll just leave it empty to indicate it's not reversible
-- In practice, you would want to track which assignments were backfilled
-- and only remove those, but for simplicity we'll leave this empty

