package routes

import (
	"github.com/TitusW/productAPI/controllers"
	"github.com/gorilla/mux"
)

func ProductRoutes(incomingRoutes *mux.Router) {
	incomingRoutes.HandleFunc("/products", controllers.GetProducts).Methods("GET")
	incomingRoutes.HandleFunc("/products/{id}", controllers.GetProductById).Methods("GET")
	incomingRoutes.HandleFunc("/products", controllers.CreateProduct).Methods("POST")
	incomingRoutes.HandleFunc("/products/{id}", controllers.UpdateProduct).Methods("PATCH")
}
