package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/rs/cors"

	_ "github.com/jinzhu/gorm/dialects/mysql"    //mysql database driver
	_ "github.com/jinzhu/gorm/dialects/postgres" //postgres database driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"   // sqlite database driver
	"github.com/nmelhado/smartmail-api/api/models"
)

// Server creates a domain that is used for all API endpoints
type Server struct {
	DB     *gorm.DB
	Router *mux.Router
}

// Initialize starts the DB connection
func (server *Server) Initialize(Dbdriver, DbUser, DbPassword, DbPort, DbHost, DbName string) {

	var err error

	DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", DbHost, DbPort, DbUser, DbName, DbPassword)
	server.DB, err = gorm.Open(Dbdriver, DBURL)
	if err != nil {
		fmt.Printf("Cannot connect to %s database", Dbdriver)
		log.Fatal("This is the error:", err)
	} else {
		fmt.Printf("We are connected to the %s database", Dbdriver)
	}

	server.DB.Debug().AutoMigrate(&models.User{}, &models.Address{}, &models.AddressAssignment{}) //database migration

	server.Router = mux.NewRouter()

	server.initializeRoutes()
}

// Run alerts that the server is up and running
func (server *Server) Run(addr string) {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // All origins
		AllowedHeaders:   []string{"*"},                     // All headers
		AllowedMethods:   []string{"POST", "GET"},
		AllowCredentials: true,
		Debug:            true,
	})

	fmt.Println("Listening to port 8080")
	log.Fatal(http.ListenAndServe(addr, c.Handler(server.Router)))
}
