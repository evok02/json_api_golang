package main

import (
	"log"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewApiServer(":4000", store)
	server.Run()
}
