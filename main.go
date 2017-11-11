package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Pigmice2733/scouting-backend/server"
	"github.com/Pigmice2733/scouting-backend/server/store/sqlite3"
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

	store, err := sqlite3.NewFromFile("scouting.db")
	if err != nil {
		log.Fatalf("error creating db: %v\n", err)
	}

	server := server.New(store, os.Stdout, tbaAPIKey, environment)

	if err := server.PollTBA("2017"); err != nil {
		fmt.Printf("error polling TBA: %v\n", err)
	}

	if err := server.Run(":" + port); err != nil {
		fmt.Printf("server died with error: %v\n", err)
	}
}
