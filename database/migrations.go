package database

import (
	"database/sql"
	"fmt"
)

// RunMigrations creates necessary database tables
func RunMigrations(db *sql.DB) error {
	migrations := []string{
		createUsersTable,
		createDataTablesTable,
		createIndexes,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %v", i+1, err)
		}
	}

	return nil
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createDataTablesTable = `
CREATE TABLE IF NOT EXISTS data_tables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    table_name VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    column_count INTEGER NOT NULL,
    row_count INTEGER NOT NULL,
    table_schema JSONB NOT NULL,
    physical_table_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_data_tables_user_id ON data_tables(user_id);
CREATE INDEX IF NOT EXISTS idx_data_tables_created_at ON data_tables(created_at);`