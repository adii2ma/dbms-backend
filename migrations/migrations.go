package migrations

import "github.com/uptrace/bun/migrate"

// Migrations registry shared across the application.
var Migrations = migrate.NewMigrations()
