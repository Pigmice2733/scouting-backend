package main

import (
	"fmt"
	"os"
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

	server := New("scouting.db", os.Stdout, environment)
	server.PollTBA("2017", tbaAPIKey)

	server.Run(":" + port)
}
