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
	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "talesmud.db"
	}
	fmt.Printf("SQLite database: %v\n", sqlitePath)

	server := server.NewApp()
	server.Run()
}
