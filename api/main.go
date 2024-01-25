// This is to Indicate that this file is part of the main package.
package handler

// Importing necessary packages for the server functionality.
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// Importing the Gorilla Mux router package to handle HTTP requests.
	// Importing the CORS package to handle Cross-Origin Resource Sharing.
	// Basically i had an issue with the Cross-Origin Resource Sharing policy, so i had to use this package to handle it.
)

type User struct {
	ID          string `json:"id" bson:"_id,omitempty"`
	Username    string `json:"username" bson:"username"`
	FullName    string `json:"fullName" bson:"fullName"`
	Email       string `json:"email" bson:"email"`
	Gender      string `json:"gender" bson:"gender"`
	BirthDate   string `json:"birthDate" bson:"birthDate"`
	PhoneNumber string `json:"phoneNumber" bson:"phoneNumber"`
}

var client *mongo.Client

// Initialization function that is executed once when the program starts.
func init() {
	// Load environment variables from .env.local
	er := godotenv.Load(".env.local")
	if er != nil {
		fmt.Println("Error loading .env.local file")
	}

	// Use the MongoDB URI from the environment variable
	mongoURI := os.Getenv("MONGODB_URI")

	clientOptions := options.Client().ApplyURI(mongoURI)
	var err error
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Connected to MongoDB")

	// Automatically create an index for the "username" field
	indexOptions := options.Index().SetUnique(true)
	usernameIndex := mongo.IndexModel{
		Keys:    bson.M{"username": 1},
		Options: indexOptions,
	}
	userCollection := client.Database("user-management-cluster").Collection("users")
	_, err = userCollection.Indexes().CreateOne(context.Background(), usernameIndex)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// Handler function to retrieve all users from the database.
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	userCollection := client.Database("user-management-cluster").Collection("users")
	cursor, err := userCollection.Find(context.Background(), bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	defer cursor.Close(context.Background())

	var users []User
	err = cursor.All(context.Background(), &users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	fmt.Println("Retrieved Users:", users)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

// Handler function to retrieve a user by ID from the database.
func GetUserByID(w http.ResponseWriter, r *http.Request) {
	// Extracting parameters from the request URL, including the user ID.
	params := mux.Vars(r)
	userID := params["id"]

	// Convert the string user ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	// Query the database to retrieve a user by their ObjectID.
	userCollection := client.Database("user-management-cluster").Collection("users")
	var user User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Println(err)
		return
	}

	// Write the user data to the response.
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// Handler function to create a new user in the database.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	userCollection := client.Database("user-management-cluster").Collection("users")
	_, err = userCollection.InsertOne(context.Background(), user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Handler function to update a user in the database by ID.
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Extracting parameters (in this case, the "id" parameter) from the request URL.
	params := mux.Vars(r)

	// Convert the string user ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	// Declaring a variable to hold the user object.
	var updatedUser User

	// Decoding the JSON request body into the user object.
	err = json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	// Querying the database to retrieve a user by their ObjectID.
	userCollection := client.Database("user-management-cluster").Collection("users")
	result, err := userCollection.UpdateOne(context.Background(), bson.M{"_id": objectID}, bson.M{"$set": updatedUser})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	// Checking if the update operation affected any rows (user not found).
	if result.ModifiedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Setting the HTTP response status code to 200 (OK).
	w.WriteHeader(http.StatusOK)

	// Encoding the updated user data as JSON and writing it to the response.
	json.NewEncoder(w).Encode(&updatedUser)
}

// Handler function to delete a user from the database by ID.
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extracting parameters (in this case, the "id" parameter) from the request URL.
	params := mux.Vars(r)

	// Convert the string user ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	// Deleting the user record from the database by ObjectID.
	userCollection := client.Database("user-management-cluster").Collection("users")
	result, err := userCollection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	// Checking if the delete operation affected any rows (user not found).
	if result.DeletedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Setting the HTTP response status code to 204 (No Content) as the resource is successfully deleted.
	w.WriteHeader(http.StatusNoContent)
}

func Main() {
	// Creating a new Gorilla Mux router.
	router := mux.NewRouter()

	// Defining HTTP routes and their corresponding handler functions.
	router.HandleFunc("/users", GetAllUsers).Methods("GET")        // Route to get all users.
	router.HandleFunc("/users/{id}", GetUserByID).Methods("GET")   // Route to get a user by ID.
	router.HandleFunc("/users", CreateUser).Methods("POST")        // Route to create a new user.
	router.HandleFunc("/users/{id}", UpdateUser).Methods("PUT")    // Route to update a user by ID.
	router.HandleFunc("/users/{id}", DeleteUser).Methods("DELETE") // Route to delete a user by ID.

	// Printing a message indicating that the server is running on port 8000, (for me to check).
	fmt.Println("Server running on port 8000")

	// Starting the HTTP server on port 8000 with the CORS-wrapped router.
	http.ListenAndServe(":8000", router)
}
