package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/TitusW/productAPI/middleware"
	"github.com/TitusW/productAPI/routes"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	router := mux.NewRouter()

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	routes.UserRoutes(router)
	productRoutes := router.PathPrefix("/products").Subrouter()
	productRoutes.Use(middleware.Authentication)
	routes.ProductRoutes(router)
	fmt.Println("[Go-Debug || GorillaMux] Listening on port :" + port)
	http.ListenAndServe(":"+port, router)
}
