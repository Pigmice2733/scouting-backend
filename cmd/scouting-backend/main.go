package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Pigmice2733/scouting-backend/internal/server"
	"github.com/Pigmice2733/scouting-backend/internal/store/postgres"
)

// ENVIRONMENT VARIABLES:
// PG_USER: postgres user
// PG_PASS: postgres password
// PG_HOST: postgres host address
// PG_PORT: postgres port
// PG_DB_NAME: postgres database name
// PG_SSL_MODE: postgres ssl mode
// PG_MAX_CONNECTIONS: postgres maximum connections
// TBA_API_KEY: the blue alliance api key
// PORT: port to listen on

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
		fmt.Printf("unable to connect to postgres server with options: %v\n", err)
		os.Exit(1)
	}

	maxConnections, err := strconv.Atoi(os.Getenv("PG_MAX_CONNECTIONS"))
	if err != nil {
		fmt.Printf("unable to parse 'PG_MAX_CONNECTIONS': %v\n", err)
		os.Exit(1)
	}

	server, err := server.New(store, os.Stdout, os.Getenv("TBA_API_KEY"), maxConnections)
	if err != nil {
		fmt.Printf("unable to create server: %v\n", err)
		os.Exit(1)
	}

	if err := server.Run(":" + os.Getenv("PORT")); err != nil {
		fmt.Printf("unable to start server: %v\n", err)
		os.Exit(1)
	}
}
