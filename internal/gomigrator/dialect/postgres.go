package dialect

import "fmt"

type Postgres struct{}

var _ QueryManager = (*Postgres)(nil)

func (p *Postgres) MigrationsTableExists(schemaName, tableName string) string {
	q := `SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = '%s' AND tablename  = '%s')`
	return fmt.Sprintf(q, schemaName, tableName)
}

func (p *Postgres) CreateMigrationsTable(schemaName, tableName string) string {
	q := `CREATE TABLE %s.%s (
		id bigint PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		applied_at timestamp NOT NULL
	)`
	return fmt.Sprintf(q, schemaName, tableName)
}

func (p *Postgres) InsertMigration(schemaName, tableName string) string {
	q := `INSERT INTO %s.%s (id, name, applied_at) VALUES ($1, $2, now())`
	return fmt.Sprintf(q, schemaName, tableName)
}

func (p *Postgres) DeleteMigration(schemaName, tableName string) string {
	q := `DELETE FROM %s.%s WHERE id = $1`
	return fmt.Sprintf(q, schemaName, tableName)
}

func (p *Postgres) ListMigrations(schemaName, tableName string) string {
	q := `SELECT id, name, applied_at FROM %s.%s ORDER BY id ASC`
	return fmt.Sprintf(q, schemaName, tableName)
}

func (p *Postgres) SelectLastMigrationID(schemaName, tableName string) string {
	q := `SELECT id FROM %s.%s ORDER BY id DESC LIMIT 1`
	return fmt.Sprintf(q, schemaName, tableName)
}
