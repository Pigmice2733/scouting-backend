package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Pigmice2733/scouting-backend/internal/server"
	"github.com/Pigmice2733/scouting-backend/internal/store/postgres"
)

func main() {
	port, err := strconv.Atoi(os.Getenv("PG_PORT"))
	if err != nil {
		fmt.Printf("unable to parse 'PG_PORT': %v\n", err)
		os.Exit(1)
	}

	store, err := postgres.NewFromOptions(postgres.Options{
		User:             os.Getenv("PG_USER"),
		Pass:             os.Getenv("PG_PASS"),
		Host:             os.Getenv("PG_HOST"),
		Port:             port,
		DBName:           os.Getenv("PG_DB_NAME"),
		SSLMode:          os.Getenv("PG_SSL_MODE"),
		StatementTimeout: 5000,
	})
	if err != nil {
		fmt.Printf("unable to connect to postgres server: %v\n", err)
		os.Exit(1)
	}

	schemaPath := "./report.schema"
	if envSchemaPath, ok := os.LookupEnv("SCHEMA_PATH"); ok {
		schemaPath = envSchemaPath
	}

	server, err := server.New(
		store, os.Stdout, os.Getenv("TBA_API_KEY"), schemaPath,
		os.Getenv("CERT_FILE"), os.Getenv("KEY_FILE"))
	if err != nil {
		fmt.Printf("unable to create server: %v\n", err)
		os.Exit(1)
	}

	if err := server.Run(os.Getenv("HTTP_ADDR"), os.Getenv("HTTPS_ADDR")); err != nil {
		fmt.Printf("unable to start server: %v\n", err)
		os.Exit(1)
	}
}
