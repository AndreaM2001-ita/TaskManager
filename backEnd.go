package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type task struct {
	ID          primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	CustomID    string             `json:"customId" bson:"customId"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	DateCreated string             `json:"dateCreated" bson:"dateCreated"`
}

var client *mongo.Client             //varibales for mongo db client
var taskCollection *mongo.Collection //varibales for mongo db

func initMongoDB() {
	// Set timeout for db in case its not present
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() //if somethign goes wrong delete conenction

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}

	// make sure conenction is ok
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping error: %v", err)
	}

	log.Println("Connected to MongoDB")

	// access db
	taskCollection = client.Database("TaskStorage").Collection("Tasks")
}

func getAllTasks(w http.ResponseWriter) {
	cursor, err := taskCollection.Find(context.Background(), bson.D{}) //get task collection by initialising cursor through find()
	if err != nil {
		http.Error(w, "Error fetching tasks", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background()) //in case something goes wrong  close the sursor

	var tasks []task

	if err := cursor.All(context.Background(), &tasks); err != nil {
		http.Error(w, "Error decoding tasks", http.StatusInternalServerError)
		return
	}
	fmt.Println("Fetched tasks from database:", tasks)
	w.Header().Set("Content-Type", "application/json") // set header of response
	json.NewEncoder(w).Encode(tasks)                   //encode the tasks info in json to be sent to frontend through body repsonse
}
func saveToDatabase(task task) {
	//create timeout for database ops
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Prepare the task for insertion into MongoDB
	newTask := bson.D{
		{Key: "customId", Value: task.CustomID},
		{Key: "name", Value: task.Name},
		{Key: "description", Value: task.Description},
		{Key: "dateCreated", Value: task.DateCreated},
	}

	// put the task int MongoDB
	result, err := taskCollection.InsertOne(ctx, newTask)
	if err != nil {
		log.Printf("Error inserting task into database: %v", err)
		return
	}

	task.ID = result.InsertedID.(primitive.ObjectID)

	log.Printf("Task successfully saved to database: %+v", task)
}

func uploadNewTask(w http.ResponseWriter, r *http.Request) {
	var task task
	err := json.NewDecoder(r.Body).Decode(&task) // Decode the incoming JSON into the task struct
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	//print outto verify what was recived
	log.Printf("Task received: %+v", task)

	go saveToDatabase(task)

	// respond with the received task
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task) // Send back the task to the frontend
}
func deleteTask(w http.ResponseWriter, customID string) {
	// fidn task object by custom id
	log.Printf("Task to delete: %+v", customID)
	//try delete task
	result, err := taskCollection.DeleteOne(context.Background(), bson.M{"customId": customID})
	if err != nil {
		http.Error(w, "Error deleting task", http.StatusInternalServerError)
		return
	}

	//check it was dleeted fully
	if result.DeletedCount == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task deleted successfully")) //rutrun 200
}
func updateTask(w http.ResponseWriter, r *http.Request, customID string) {
	var updatedTask task

	// get new values nd put them into the task struct
	//decode body request
	if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	//create new struct
	update := bson.M{
		"$set": bson.M{
			"name":        updatedTask.Name,
			"description": updatedTask.Description,
			"dateCreated": updatedTask.DateCreated,
		},
	}

	//update
	_, err := taskCollection.UpdateOne(
		context.Background(),
		bson.M{"customId": customID},
		update,
	)

	if err != nil {
		http.Error(w, "Error updating task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedTask); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
func handleTasks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customID, hasCustomID := vars["customId"]

	switch r.Method {
	case http.MethodPost: //POST
		uploadNewTask(w, r)
	case http.MethodGet: //GET
		getAllTasks(w)
	case http.MethodDelete: //DELETE
		if hasCustomID {
			deleteTask(w, customID)
		} else {
			http.Error(w, "Missing custom ID for delete", http.StatusBadRequest)
		}
	case http.MethodPut: //UPDATE
		if hasCustomID {
			updateTask(w, r, customID)
		} else {
			http.Error(w, "Missing custom ID for update", http.StatusBadRequest)
		}
	default: //NOT FOUND
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func main() {
	initMongoDB()
	defer func() { // defer calls this fucntion once  main is concluded,
		// so that connection with DB is alwys elegantly closed
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("MongoDB disconnect error: %v", err)
		}
	}()

	//Routing of restful API CRUD operations
	router := mux.NewRouter()
	router.HandleFunc("/api/tasks", handleTasks).Methods("GET", "POST")
	router.HandleFunc("/api/tasks/{customId}", handleTasks).Methods("PUT", "DELETE")

	log.Println("server running on localhost:8080")

	// apply CORS middleware to allow entries from the same localhost
	err := http.ListenAndServe(":8080",
		handlers.CORS(
			handlers.AllowedOrigins([]string{"http://localhost:3000",
				"http://127.0.0.1:3000"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
			handlers.AllowedHeaders([]string{"Content-Type"}),
		)(router))

	if err != nil {
		log.Fatalf("server cannot start: %v", err)
	}
}
