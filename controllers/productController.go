package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/TitusW/productAPI/database"
	"github.com/TitusW/productAPI/models"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	productCollection *mongo.Collection = database.OpenCollection(database.Client, "Product")
)

func GetProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	var Products []models.Product

	result, err := productCollection.Find(context.TODO(), bson.M{})
	defer cancel()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	if err = result.All(ctx, &Products); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Products)
}

func GetProductById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	var Product models.Product
	id := mux.Vars(r)["id"]

	if err := productCollection.FindOne(ctx, bson.M{"product_id": id}).Decode(&Product); err != nil {
		defer cancel()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	defer cancel()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Product)
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	var Product models.Product

	if err := json.NewDecoder(r.Body).Decode(&Product); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	Product.ID = primitive.NewObjectID()
	Product.Product_id = Product.ID.Hex()
	Product.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	Product.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	result, insertErr := productCollection.InsertOne(ctx, Product)
	if insertErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(insertErr)
		return
	}

	defer cancel()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	var Product models.Product
	id := mux.Vars(r)["id"]

	if err := json.NewDecoder(r.Body).Decode(&Product); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	var updateObj primitive.D

	if Product.Name != nil {
		updateObj = append(updateObj, bson.E{"name", Product.Name})
	}

	if Product.Product_image != nil {
		updateObj = append(updateObj, bson.E{"product_image", Product.Product_image})
	}

	Product.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", Product.Updated_at})

	upsert := true
	filter := bson.M{"product_id": id}

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	result, err := productCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	defer cancel()
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(result)
}
