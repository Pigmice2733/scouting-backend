package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Pigmice2733/scouting-backend/server"
	"github.com/Pigmice2733/scouting-backend/server/store/postgres"
)

func main() {
	log.Println("Starting backend")
	tbaAPIKey, valid := os.LookupEnv("TBA_API_KEY")
	if !valid {
		log.Println("Valid TBA API key must be provided in environment variable 'TBA_API_KEY'")
	}

	port, valid := os.LookupEnv("PORT")
	if !valid {
		port = "8080"
	}

	environment, valid := os.LookupEnv("ENVIRONMENT")
	if !valid {
		environment = "dev"
	}

	postgresOptions, err := dbOptions(environment)
	if err != nil {
		log.Printf("error getting connection info for PostgreSQL: '%v'\n", err)
		return
	}

	store, err := postgres.NewFromOptions(postgresOptions)
	if err != nil {
		log.Printf("error: connecting to database: %v\n", err)
		return
	}
	defer store.Close()
	log.Println("Connected to database")

	server, err := server.New(store, os.Stdout, tbaAPIKey, environment)
	if err != nil {
		log.Printf("error creating server: %v\n", err)
		return
	}
	log.Println("Created server")

	if err := server.PollTBA("2017"); err != nil {
		log.Printf("error polling TBA: %v\n", err)
	}

	if err := server.Run(":" + port); err != nil {
		store.Close()
		log.Fatalf("server died with error: %v\n", err)
	}
}

func dbOptions(environment string) (postgres.Options, error) {
	maxConnections, err := strconv.Atoi(os.Getenv("MAX_DB_CONNECTIONS"))
	if err != nil {
		return postgres.Options{}, fmt.Errorf("error parsing MAX_DB_CONNECTIONS: %v\n", err)
	}

	const defaultMaxConnections = 32
	if maxConnections < 1 {
		log.Printf("MAX_DB_CONNECTIONS must be >= 1, defaulting to %d\n", defaultMaxConnections)
		maxConnections = defaultMaxConnections
	}

	port, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if err != nil {
		return postgres.Options{}, fmt.Errorf("error parsing POSTGRES_PORT: %v\n", err)
	}

	return postgres.Options{
		User:             os.Getenv("POSTGRES_USER"),
		Pass:             os.Getenv("POSTGRES_PASS"),
		Host:             os.Getenv("POSTGRES_HOST"),
		Port:             port,
		DBName:           os.Getenv("POSTGRES_DBNAME"),
		SSLMode:          os.Getenv("POSTGRES_SSLMODE"),
		MaxConnections:   maxConnections,
		StatementTimeout: 5000,
	}, nil
}
