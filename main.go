package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

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
	if strings.ToLower(environment) == "staging" {
		rawConnectionURL, valid := os.LookupEnv("POSTGRES_URL")
		if !valid {
			return postgres.Options{}, fmt.Errorf("staging missing valid POSTGRES_URL for PostgreSQL connection")
		}

		connectionURL, err := url.Parse(rawConnectionURL)
		if err != nil {
			return postgres.Options{}, fmt.Errorf("staging environment variable POSTGRES_URL does not contain valid connection info")
		}

		user := connectionURL.User.Username()
		password, _ := connectionURL.User.Password()
		host, portStr, err := net.SplitHostPort(connectionURL.Host)
		dbName := connectionURL.Path[1:]
		sslMode, validSSLMode := os.LookupEnv("POSTGRES_SSL")
		if !validSSLMode {
			return postgres.Options{}, fmt.Errorf("staging missing environment variable POSTGRES_SSL specifying SSL mode")
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			return postgres.Options{}, fmt.Errorf("staging POSTGRES_URL must contain valid port number")
		}

		maxConnections, err := getMaxConnections()
		if err != nil {
			return postgres.Options{}, err
		}

		return postgres.Options{User: user, Pass: password, Host: host, Port: port, DBName: dbName, SSLMode: sslMode, MaxConnections: maxConnections, StatementTimeout: 5000}, nil
	} else if strings.ToLower(environment) == "production" {
		user, validUser := os.LookupEnv("POSTGRES_USER")
		password, validPass := os.LookupEnv("POSTGRES_PASSWORD")
		host, validHost := os.LookupEnv("DATABASE_URL")
		portStr, validPort := os.LookupEnv("POSTGRES_PORT")
		dbName, validDBName := os.LookupEnv("POSTGRES_DB")
		sslMode, validSSLMode := os.LookupEnv("POSTGRES_SSL")

		if !validUser || !validPass || !validHost || !validPort || !validDBName || !validSSLMode {
			return postgres.Options{}, fmt.Errorf("production environment missing required environment variables for PostgreSQL database")
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return postgres.Options{}, fmt.Errorf("production environment must contain valid 'POSTGRES_PORT' number")
		}

		maxConnections, err := getMaxConnections()
		if err != nil {
			return postgres.Options{}, err
		}

		return postgres.Options{User: user, Pass: password, Host: host, Port: port, DBName: dbName, SSLMode: sslMode, MaxConnections: maxConnections, StatementTimeout: 5000}, nil
	}
	return postgres.Options{}, fmt.Errorf("environment '%s' not supported", environment)
}

func getMaxConnections() (int, error) {
	maxConnectionsStr, valid := os.LookupEnv("MAX_DB_CONNECTIONS")
	if valid {
		maxConnections, err := strconv.Atoi(maxConnectionsStr)
		if err != nil {
			return maxConnections, fmt.Errorf("MAX_DB_CONNECTIONS environment variable contains invalid connection number")
		}
		return maxConnections, nil
	}
	defaultMax := 60
	log.Printf("MAX_DB_CONNECTIONS should contain maximum database connections. Using default %v\n", defaultMax)
	return defaultMax, nil
}
