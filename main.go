package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dakamakat/RSS/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

    _ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	godotenv.Load(".env")

	portString := os.Getenv("PORT")

	if portString == "" {
		log.Fatal("Port is not found in the environment")
	}

	dbUrl := os.Getenv("DB_URL")

	if dbUrl == "" {
		log.Fatal("DB Url is not found in the environment")
	}

	conn, err := sql.Open("postgres", dbUrl)

	if err != nil {
		log.Fatal("Cant connect to Database")
	}

    queries := database.New(conn)

    apiCfg := apiConfig{
        DB: queries,
    }

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()

	v1Router.Get("/health", handlerReadiness)
	v1Router.Get("/error", handlerError)

	v1Router.Get("/users", apiCfg.handlerGetUser)
	v1Router.Post("/users", apiCfg.handlerCreateUser)

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	log.Printf("Server starting on port %v", portString)

	srv_err := srv.ListenAndServe()

	if srv_err != nil {
		log.Fatal(err)
	}

	fmt.Println("Port: ", portString)
}
