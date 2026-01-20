package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
	"github.com/talesmud/talesmud/pkg/server"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file")
	}

	logLevel := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithField("LOG_LEVEL", logLevel).Warn("Invalid LOG_LEVEL, defaulting to info")
		level = log.InfoLevel
	}
	log.SetLevel(level)

	fmt.Println("Starting tales server...")
	driver := strings.ToLower(strings.TrimSpace(os.Getenv("DB_DRIVER")))
	if driver == "" && strings.TrimSpace(os.Getenv("SQLITE_PATH")) != "" {
		driver = "sqlite"
	}
	if driver == "" {
		driver = "mongo"
	}
	fmt.Printf("db driver %v\n", driver)
	if driver == "sqlite" {
		fmt.Printf("sqlite path %v\n", os.Getenv("SQLITE_PATH"))
	} else {
		fmt.Printf("mongo connection string %v\n", os.Getenv("MONGODB_CONNECTION_STRING"))
		fmt.Printf("mongo database %v\n", os.Getenv("MONGODB_DATABASE"))
	}

	server := server.NewApp()
	server.Run()
}
