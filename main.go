package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Pigmice2733/scouting-backend/server"
	"github.com/Pigmice2733/scouting-backend/server/store/postgres"
)

func main() {
	tbaAPIKey, valid := os.LookupEnv("TBA_API_KEY")
	if !valid {
		fmt.Println("Valid TBA API key must be provided in environment variable 'TBA_API_KEY'")
	}

	port, valid := os.LookupEnv("PORT")
	if !valid {
		port = "8080"
	}

	fmt.Println("Using port " + port)

	environment, valid := os.LookupEnv("ENVIRONMENT")
	if !valid {
		environment = "dev"
	}

	store, err := postgres.NewFromOptions(postgres.Options{User: os.Getenv("POSTGRES_1_ENV_POSTGRES_USER"), Pass: os.Getenv("POSTGRES_1_ENV_POSTGRES_PASSWORD"), Host: os.Getenv("POSTGRES_1_PORT_5432_TCP_ADDR"), Port: 5432, DBName: os.Getenv("POSTGRES_1_ENV_POSTGRES_DB"), SSLMode: "disable", StatementTimeout: 5000})
	if err != nil {
		log.Fatalf("error: creating database: %v\n", err)
	}

	server, err := server.New(store, os.Stdout, tbaAPIKey, environment)
	if err != nil {
		fmt.Printf("error starting server: %v\n", err)
		return
	}

	if err := server.PollTBA("2017"); err != nil {
		fmt.Printf("error polling TBA: %v\n", err)
	}

	if err := server.Run(":" + port); err != nil {
		fmt.Printf("server died with error: %v\n", err)
	}
}
