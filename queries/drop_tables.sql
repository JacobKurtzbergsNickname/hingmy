-- Drop all tables (in correct order to respect foreign key constraints)

-- Drop dependent tables first
DROP TABLE IF EXISTS notes;
DROP TABLE IF EXISTS tag_entities;

-- Drop main tables
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS todos;

-- Note: SQLite doesn't enforce foreign key constraints by default,
-- but it's good practice to drop in the correct order