package main

import (
	"api-test/database"
	"api-test/server"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Keine .env Datei gefunden")
	}

	database.InitDB()
	server.NewServer()
}
