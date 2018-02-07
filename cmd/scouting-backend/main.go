package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Pigmice2733/scouting-backend/internal/server"
	"github.com/Pigmice2733/scouting-backend/internal/store/postgres"
	"github.com/Pigmice2733/scouting-backend/internal/tba/api"
)

const tbaURL = "http://www.thebluealliance.com/api/v3"

func main() {
	port, err := strconv.Atoi(os.Getenv("PG_PORT"))
	if err != nil {
		port = 5432
		fmt.Printf("PG_PORT defaulted to: %d\n", port)
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

	consumer := api.New(tbaURL, os.Getenv("TBA_API_KEY"))

	schemaPath := "./report.schema"
	if envSchemaPath, ok := os.LookupEnv("SCHEMA_PATH"); ok {
		schemaPath = envSchemaPath
	}

	origin := "*"
	if envOrigin, ok := os.LookupEnv("ORIGIN"); ok {
		origin = envOrigin
	}

	server, err := server.New(
		store, consumer, os.Stdout, origin, schemaPath,
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
