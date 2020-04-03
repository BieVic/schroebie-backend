package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"net/http"
)

type Painting struct {
	Id 		primitive.ObjectID 	`bson:"_id,omitempty"`
	Binary 	[]byte 				`bson:"binary"`
	Title 	string 				`bson:"title"`
	Artist 	string 				`bson:"artist"`
	Year 	string 				`bson:"year"`
	Size 	string 				`bson:"size"`
	Sold	bool				`bson:"sold"`
}

var collection *mongo.Collection

func initMongoDB() *mongo.Client {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://schroebiedb:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	return client
}

func save(p Painting) {
	_, err := collection.InsertOne(context.TODO(), p)
	if err != nil {
		log.Fatal(err)
	}
}

func getAll() []*Painting {
	findOptions := options.Find()
	var results []*Painting
	cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(context.TODO()) {
		var elem Painting
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, &elem)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	// Close the cursor once finished
	cur.Close(context.TODO())
	return results
}

func uploadPainting(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("picture")
	if err != nil {
		fmt.Println("Error Retrieving the file")
		fmt.Println(err)
		return
	}
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	title := r.FormValue("title")
	size := r.FormValue("size")
	year := r.FormValue("year")
	artist := r.FormValue("artist")
	sold := r.FormValue("sold") == "Sold"
	painting := Painting{
		Id:     primitive.ObjectID{},
		Binary: buf,
		Title:  title,
		Artist: artist,
		Year:   year,
		Size:   size,
		Sold:	sold,
	}

	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)

	save(painting)
}

func landingPage(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Host)
	paintings := getAll()
	js, err := json.Marshal(paintings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func defineRoutes() {
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/", fs)
	http.HandleFunc("/gallery", landingPage)
	http.HandleFunc("/upload", uploadPainting)
}

func main() {
	client := initMongoDB()
	collection = client.Database("paintings").Collection("pwolff")
	defer client.Disconnect(context.TODO())

	defineRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
