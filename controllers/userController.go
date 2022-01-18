package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/TitusW/productAPI/database"
	"github.com/TitusW/productAPI/helpers"
	"github.com/TitusW/productAPI/models"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	userCollection *mongo.Collection = database.OpenCollection(database.Client, "User")
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := helpers.CheckUserType(w, r, "ADMIN"); err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	urlParams := r.URL.Query()

	recordPerPage, err := strconv.Atoi(urlParams["recordPerPage"][0])
	if err != nil || recordPerPage < 1 {
		recordPerPage = 10
	}
	page, err := strconv.Atoi(urlParams["page"][0])
	if err != nil || page < 1 {
		page = 1
	}

	startIndex := (page - 1) * recordPerPage

	matchStage := bson.D{{"$match", bson.D{{}}}}
	groupStage := bson.D{{"$group", bson.D{
		{"_id", bson.D{{"_id", "null"}}},
		{"total_count", bson.D{{"$sum", 1}}},
		{"data", bson.D{{"$push", "$$ROOT"}}},
	}}}
	projectStage := bson.D{
		{"$project", bson.D{
			{"_id", 0},
			{"total_count", 1},
			{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
		}},
	}

	result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage, groupStage, projectStage,
	})
	defer cancel()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	var Users []bson.M
	if err = result.All(ctx, &Users); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Users)
}

func GetUserById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	var User models.User
	id := mux.Vars(r)["id"]

	if err := helpers.MatchUserTypeToUid(w, r, id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if err := userCollection.FindOne(ctx, bson.M{"user_id": id}).Decode(&User); err != nil {
		defer cancel()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	defer cancel()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(User)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	var User models.User

	if err := json.NewDecoder(r.Body).Decode(&User); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		json.NewEncoder(w).Encode("Error signing up")
		return
	}

	password := HashPassword(*User.Password)
	User.Password = &password
	count, err := userCollection.CountDocuments(ctx, bson.M{"email": User.Email})
	defer cancel()
	if err != nil {
		log.Panic(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	if count > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("This email already exist")
		return
	}

	User.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	User.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	User.ID = primitive.NewObjectID()
	User.User_id = User.ID.Hex()
	token, refreshToken, _ := helpers.GenerateAllTokens(*User.Email, *User.First_name, *User.Last_name, *&User.User_id)
	User.Token = &token
	User.Refresh_Token = &refreshToken

	result, insertErr := userCollection.InsertOne(ctx, User)
	if insertErr != nil {
		msg := fmt.Sprintf("User item was not created")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	var User models.User
	var FoundUser models.User

	if err := json.NewDecoder(r.Body).Decode(&User); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	err := userCollection.FindOne(ctx, bson.M{"email": User.Email}).Decode(&FoundUser)
	defer cancel()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	passwordIsValid, msg := VerifyPassword(*User.Password, *FoundUser.Password)
	if passwordIsValid != true {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(msg)
		return
	}

	if FoundUser.Email == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("User not found")
		return
	}

	token, refreshToken, _ := helpers.GenerateAllTokens(*FoundUser.Email, *FoundUser.First_name, *FoundUser.Last_name, *&FoundUser.User_id)
	helpers.UpdateAllTokens(token, refreshToken, FoundUser.User_id)
	err = userCollection.FindOne(ctx, bson.M{"user_id": FoundUser.User_id}).Decode(&FoundUser)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(FoundUser)
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("email or password is incorrect")
		check = true
	}

	return check, msg
}
