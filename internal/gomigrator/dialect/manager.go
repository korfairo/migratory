package dialect

type QueryManager interface {
	MigrationsTableExists(schemaName, tableName string) string
	CreateMigrationsTable(schemaName, tableName string) string
	InsertMigration(schemaName, tableName string) string
	DeleteMigration(schemaName, tableName string) string
	ListMigrations(schemaName, tableName string) string
	SelectLastMigrationID(schemaName, tableName string) string
}
