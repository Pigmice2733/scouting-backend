package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config file structure
type Config struct {
	TBAApikey string `json:"tbaApikey"`
}

func main() {
	config := loadConfiguration("config.json")

	server := Server{}
	server.Initialize("scouting.db")
	server.PollTBA("2017", config.TBAApikey)

	server.Run(":8080")
}

func loadConfiguration(file string) Config {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		fmt.Println(err.Error())
	}
	return config
}
