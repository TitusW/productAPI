package routes

import (
	"github.com/TitusW/productAPI/controllers"
	"github.com/gorilla/mux"
)

func UserRoutes(incomingRoutes *mux.Router) {
	incomingRoutes.HandleFunc("/users", controllers.GetUsers).Methods("GET")
	incomingRoutes.HandleFunc("/users/{id}", controllers.GetUserById).Methods("GET")
	incomingRoutes.HandleFunc("/users/signup", controllers.Signup).Methods("POST")
	incomingRoutes.HandleFunc("/users/login", controllers.Login).Methods("POST")
}
