package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"os"

	"github.com/joho/godotenv"
	"github.com/talesmud/talesmud/pkg/server"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file")
	}

	fmt.Println("Starting tales server...")

	fmt.Printf("mongo connection string %v\n", os.Getenv("MONGODB_CONNECTION_STRING"))
	fmt.Printf("mongo database %v\n", os.Getenv("MONGODB_DATABASE"))

	server := server.NewApp()
	server.Run()
}
