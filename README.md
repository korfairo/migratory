# migratory
## minimalistic database migration library and CLI tool

Supports PostgreSQL. Works as a library (go package) and CLI tool.

## As library

Works with **database/sql** standard package.

```shell
go get github.com/korfairo/migratory@70c1c710b676c66878dbe92fa89864677a0bc541
```

### API

Register your `.go` migrations with functions:
```
- func AddMigration(up, down GoMigrateFn)
- func AddMigrationNoTx(up, down GoMigrateNoTxFn)
````
You can also use `.sql` migrations. Set directory with migrations files:

```
- func SetSQLDirectory(path string)
```

Or use `OptionsFunc` `WithSQLMigrationDir(d string)` in the next commands.

Manage your migrations with functions:

```
- func Up(db *sql.DB, opts ...OptionsFunc) (n int, err error)
- func Down(db *sql.DB, opts ...OptionsFunc) error
- func Redo(db *sql.DB, opts ...OptionsFunc) error
- func GetStatus(db *sql.DB, opts ...OptionsFunc) ([]MigrationResult, error)
- func GetDBVersion(db *sql.DB, opts ...OptionsFunc) (int64, error)
```

Or with your context:

```
- func UpContext(ctx context.Context, db *sql.DB, opts ...OptionsFunc) (n int, err error)
- func DownContext(ctx context.Context, db *sql.DB, opts ...OptionsFunc) error
- func RedoContext(ctx context.Context, db *sql.DB, opts ...OptionsFunc) error
- func GetStatusContext(ctx context.Context, db *sql.DB, opts ...OptionsFunc) ([]MigrationResult, error)
- func GetDBVersionContext(ctx context.Context, db *sql.DB, opts ...OptionsFunc) (int64, error)
```

### As CLI tool

```shell
go install github.com/korfairo/migratory/cmd/migratory@70c1c710b676c66878dbe92fa89864677a0bc541
```

### Usage

```
Usage:
  migratory [command]

Available Commands:
  create      Creates .sql or .go migration template
  dbversion   Shows the DB version (id of the last applied migration
  down        Rollback last applied migration
  help        Help about any command
  redo        Rollbacks and applies again last migration
  status      Shows migration statuses
  up          Up all unapplied migrations

Flags:
  -c, --config string   path to yaml config
  -d, --db string       database connection string
      --dir string      directory with .sql migration files (default ".")
  -h, --help            help for migratory
  -s, --schema string   name of database schema with migrations table (default "public")
  -t, --table string    name of migrations table (default "migrations")
```

You can find information about all commands and their usage with --help or -h flag.

All commands works with config (.yml file). Create a configuration file and pass its path with the `-c ./path/` flag.

Config example:
```yaml
directory: /path/to/directory
dsn: postgres://user:password@localhost:5432/my_db
schema: public
table: migrations
```

### Create migration template
#### migratory create

```
This command creates .sql or .go file with standard migration template. 
Default directory is your current one, pass arg with -d flag to choose another.
Name of the file matches the format {id}_{name}.sql, where id is a unique number of migration.
The command writes current UTC time as a migration id, for example: 20060102150405_name.sql

Usage:
  migratory create {<name>} {sql|go} [--dir <path>] [flags]

Examples:
migratory create my_migration go
migratory create my_migration sql
migratory create my_migration sql -d ./example/migrations

Flags:
  -h, --help   help for create

Global Flags:
  -c, --config string   path to yaml config
  -d, --db string       database connection string
      --dir string      directory with .sql migration files (default ".")
  -s, --schema string   name of database schema with migrations table (default "public")
  -t, --table string    name of migrations table (default "migrations")
```

### Up all missing migrations
#### migratory up

```
The "up" command applies all unapplied migrations. 
It searches for SQL migration files in the directory that is passed as an argument, 
checks the version of the database - which is the ID of the last applied migration 
in your migrations database table - and applies any missing migrations one-by-one. 
If there are migrations in your directory with ID numbers less than the database version, 
they are considered "dirty migrations". By default, the command will return an error 
in this case, but you can ignore it with the "–force" flag to apply all missing migrations. 
Additionally, the command will create the migrations table if it does not already exist.

Usage:
  migratory up [--dir <path>] [-d <db-string>] [-s <schema>] [-t <table>] [flags]

Examples:
migratory up -c /etc/config.yml
migratory up -d postgresql://role:password@127.0.0.1:5432/database --dir example/migrations/
migratory up -d postgresql://role:password@127.0.0.1:5432/database --dir migrations/ -t my_migrations_table

Flags:
  -f, --force   ignore "dirty migrations" error
  -h, --help    help for up

Global Flags:
  -c, --config string   path to yaml config
  -d, --db string       database connection string
      --dir string      directory with .sql migration files (default ".")
  -s, --schema string   name of database schema with migrations table (default "public")
  -t, --table string    name of migrations table (default "migrations")
```

### Down last migration
#### migratory down

```
The "down" command rolls back the last applied migration.
Command creates migrations table if not exists.

Usage:
  migratory down [--dir <path>] [-d <db-string>] [-s <schema>] [-t <table>] [flags]

Examples:
migratory down -c /etc/config.yml
migratory down -d postgresql://role:password@127.0.0.1:5432/database --dir example/migrations/
migratory down -d postgresql://role:password@127.0.0.1:5432/database --dir migrations/ -t my_migrations_table

Flags:
  -h, --help   help for down

Global Flags:
  -c, --config string   path to yaml config
  -d, --db string       database connection string
      --dir string      directory with .sql migration files (default ".")
  -s, --schema string   name of database schema with migrations table (default "public")
  -t, --table string    name of migrations table (default "migrations")
```

### Redo
#### migration redo

```
The "redo" command rolls back the last applied migration, then applies it again.
Command creates migrations table if not exists.

Usage:
  migratory redo [--dir <path>] [-d <db-string>] [-s <schema>] [-t <table>] [flags]

Examples:
migratory redo -c /etc/config.yml
migratory redo -d postgresql://role:password@127.0.0.1:5432/database --dir example/migrations/
migratory redo -d postgresql://role:password@127.0.0.1:5432/database --dir migrations/ -t my_migrations_table

Flags:
  -h, --help   help for redo

Global Flags:
  -c, --config string   path to yaml config
  -d, --db string       database connection string
      --dir string      directory with .sql migration files (default ".")
  -s, --schema string   name of database schema with migrations table (default "public")
  -t, --table string    name of migrations table (default "migrations")
```

### Get database version (ID of the last applied migration)
#### migratory dbversion

```
The "dbversion" command prints the id of the last applied migration 
from migrations table in your database. Command creates migrations table if not exists.

Usage:
  migratory dbversion [-d <db-string>] [-s <schema>] [-t <table>] [flags]

Examples:
dbversion -c /etc/config.yml
dbversion -d postgresql://role:password@127.0.0.1:5432/database
dbversion -d postgresql://role:password@127.0.0.1:5432/database -s my_schema -t my_migrations_table

Flags:
  -h, --help   help for dbversion

Global Flags:
  -c, --config string   path to yaml config
  -d, --db string       database connection string
      --dir string      directory with .sql migration files (default ".")
  -s, --schema string   name of database schema with migrations table (default "public")
  -t, --table string    name of migrations table (default "migrations")

```

### Get database status (table with all applied and unapplied migrations)
#### migratory status

```
The "status" command shows table with migration statuses,
according existing migrations in your directory and in the database migrations table.
Command creates migrations table if not exists.

Usage:
  migratory status [--dir <path>] [-d <db-string>] [-s <schema>] [-t <table>] [flags]

Examples:
status -c /etc/config.yml
status -d postgresql://role:password@127.0.0.1:5432/database --dir example/migrations/
status -d postgresql://role:password@127.0.0.1:5432/database --dir migrations/ -t my_migrations_table

Flags:
  -h, --help   help for status

Global Flags:
  -c, --config string   path to yaml config
  -d, --db string       database connection string
      --dir string      directory with .sql migration files (default ".")
  -s, --schema string   name of database schema with migrations table (default "public")
  -t, --table string    name of migrations table (default "migrations")
```

Example of status table:
```
ID  Name                      Applied  Date                 
1   simple                    true     2024-02-01 14:48:11  
2   with_spaces_and_newlines  true     2024-02-01 14:48:11  
3   with_comments             true     2024-02-01 14:48:11  
4   multiple_statements       true     2024-02-01 14:48:11  
5   statement_tags            false    0001-01-01 00:00:00  
6   notransaction             false    0001-01-01 00:00:00  
7   no_down                   false    0001-01-01 00:00:00  
8   no_up                     false    0001-01-01 00:00:00 
```

