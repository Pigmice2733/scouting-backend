package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/Pigmice2733/scouting-backend/internal/store/postgres"
	_ "github.com/lib/pq"
)

const usageFormat = `Usage: %s [...table]

Environment Variables:
* PG_USER: postgresql user
* PG_PASS: postgresql password
* PG_HOST: postgresql host address
* PG_PORT: postgresql port (defaults to 5432)
* PG_DB_NAME: postgresql db name
* PG_SSL_MODE: postgresql ssl mode
`

func main() {
	port, err := strconv.Atoi(os.Getenv("PG_PORT"))
	if err != nil {
		port = 5432
	}

	options := postgres.Options{
		User:             os.Getenv("PG_USER"),
		Pass:             os.Getenv("PG_PASS"),
		Host:             os.Getenv("PG_HOST"),
		Port:             port,
		DBName:           os.Getenv("PG_DB_NAME"),
		SSLMode:          os.Getenv("PG_SSL_MODE"),
		StatementTimeout: 5000,
	}

	db, err := sql.Open("postgres", options.ConnectionInfo())
	if err != nil {
		fmt.Printf("error connecting to postgresql database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if len(os.Args) == 1 {
		fmt.Printf(usageFormat, os.Args[0])
		os.Exit(2)
	}

	for _, table := range os.Args[1:] {
		if err := archiveQuery(db, table, table+".json"); err != nil {
			fmt.Printf("error archiving table '%s': %v\n", "events", err)
		}
	}
}

const jsonArchiveFormat = `SELECT array_to_json(array_agg(%s)) FROM %s`

func archiveQuery(db *sql.DB, table, out string) error {
	query := fmt.Sprintf(jsonArchiveFormat, table, table)

	var archive string
	if err := db.QueryRow(query).Scan(&archive); err != nil {
		return err
	}

	f, err := os.OpenFile("events.json", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(archive)
	return err
}
