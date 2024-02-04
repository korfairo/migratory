### temporary migration, see #14721 ###

-- +migrate up

-- This column was added in #14721.
-- approved by team lead

ALTER TABLE users
ADD COLUMN created_at TIMESTAMP DEFAULT NOW();

-- +migrate down
ALTER TABLE users
DROP COLUMN created_at TIMESTAMP DEFAULT NOW();
