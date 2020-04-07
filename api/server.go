package api

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nmelhado/smartmail-api/api/controllers"
	// "github.com/nmelhado/smartmail-api/api/seed"
)

var server = controllers.Server{}

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("sad .env file found")
	}
}

func Run() {

	if os.Getenv("APP_ENV") != "production" {
		var err error
		err = godotenv.Load()
		if err != nil {
			log.Fatalf("Error getting env, %v\nChecking other env variables...", err)
		} else {
			fmt.Println("Getting the env values...")
		}
	}

	server.Initialize(os.Getenv("DB_DRIVER"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_HOST"), os.Getenv("PROD_DB_HOST"), os.Getenv("DB_NAME"))

	// seed.Load(server.DB)

	server.Run(":8080")

}
