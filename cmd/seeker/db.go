package main

import (
	_ "database/sql/driver"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"
)

func runDBMigration(conn *pgx.Conn) error {
	m, err := migrate.New(
		"file://migrations",
		fmt.Sprintf("%s?sslmode=disable", os.Getenv("DATABASE_URL")))

	if err != nil {
		return err
	}

	m.Up()

	return nil
}
