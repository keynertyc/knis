package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Person representa los datos de una persona
type Person struct {
	DNI             string    `json:"dni"`
	Nombre          string    `json:"nombre"`
	ApellidoPaterno string    `json:"apellido_paterno"`
	ApellidoMaterno string    `json:"apellido_materno"`
	Domicilio       struct {
		Direccion    string `json:"direccion"`
		Distrito     string `json:"distrito"`
		Provincia    string `json:"provincia"`
		Departamento string `json:"departamento"`
		Ubigeo       string `json:"ubigeo"`
	} `json:"domicilio"`
	CreatedAt time.Time `json:"createdat"`
	UpdatedAt time.Time `json:"updatedat"`
}

type Response struct {
	Success bool   `json:"success"`
	Data    Person `json:"data"`
	Source  int    `json:"source"`
}

type MongoDBClient struct {
	Client     *mongo.Client
	Collection *mongo.Collection
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Debes proporcionar un rango de DNIs, ejemplo: 40000000 49999999")
	}

	start, end := parseRange(os.Args[1], os.Args[2])
	mongoURI, dbName := getEnvVars()

	mongoClient := NewMongoDBClient(mongoURI, dbName, "personas")
	defer mongoClient.Disconnect()

	// Utilizar WaitGroup para manejar concurrencia
	var wg sync.WaitGroup

	for dni := start; dni <= end; dni++ {
		wg.Add(1)
		go func(dni int) {
			defer wg.Done()
			processDNI(dni, mongoClient.Collection)
		}(dni)
	}

	wg.Wait()
	fmt.Println("Proceso completado")
}

func parseRange(startStr, endStr string) (int, int) {
	start, err := strconv.Atoi(startStr)
	if err != nil {
		log.Fatal("Error al convertir el inicio del rango:", err)
	}
	end, err := strconv.Atoi(endStr)
	if err != nil {
		log.Fatal("Error al convertir el final del rango:", err)
	}
	return start, end
}

func getEnvVars() (string, string) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI no está configurado")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("DB_NAME no está configurado")
	}
	return mongoURI, dbName
}

func NewMongoDBClient(uri, dbName, collectionName string) *MongoDBClient {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	return &MongoDBClient{
		Client:     client,
		Collection: client.Database(dbName).Collection(collectionName),
	}
}

func (m *MongoDBClient) Disconnect() {
	if err := m.Client.Disconnect(context.TODO()); err != nil {
		log.Fatal("Error al desconectar de MongoDB:", err)
	}
}

func processDNI(dni int, collection *mongo.Collection) {
	data, err := fetchData(dni)
	if err != nil {
		log.Println("Error al obtener los datos para DNI", dni, ":", err)
		return
	}

	if data == nil {
		log.Println("Solicitud no exitosa para DNI", dni)
		return
	}

	err = upsertData(dni, data, collection)
	if err != nil {
		log.Println("Error al guardar los datos en la base de datos para DNI", dni, ":", err)
		return
	}

	fmt.Printf("Datos guardados correctamente para DNI %d\n", dni)
}

func fetchData(dni int) (*Response, error) {
	url := fmt.Sprintf("https://dniruc.apisunat.com/dni/%d", dni)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear el request: %v", err)
	}

	req.Header.Add("accept", "application/json, text/plain, */*")
	req.Header.Add("origin", "https://apisunat.com")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al hacer el request: %v", err)
	}
	defer resp.Body.Close()

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error al deserializar JSON: %v", err)
	}

	if !response.Success {
		return nil, nil
	}

	return &response, nil
}

func upsertData(dni int, response *Response, collection *mongo.Collection) error {
	currentTime := time.Now()

	filter := bson.M{"dni": response.Data.DNI}
	var existingDoc Person
	err := collection.FindOne(context.TODO(), filter).Decode(&existingDoc)

	if err == mongo.ErrNoDocuments {
		return insertNewDocument(response, collection, currentTime)
	} else if err == nil {
		return updateExistingDocument(response, collection, filter, currentTime)
	} else {
		return fmt.Errorf("error al buscar documento existente: %v", err)
	}
}

func insertNewDocument(response *Response, collection *mongo.Collection, currentTime time.Time) error {
	response.Data.CreatedAt = currentTime
	response.Data.UpdatedAt = currentTime

	doc := bson.M{
		"dni":              response.Data.DNI,
		"nombre":           response.Data.Nombre,
		"apellido_paterno": response.Data.ApellidoPaterno,
		"apellido_materno": response.Data.ApellidoMaterno,
		"domicilio":        response.Data.Domicilio,
		"source":           response.Source,
		"createdat":        response.Data.CreatedAt,
		"updatedat":        response.Data.UpdatedAt,
	}

	_, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		return fmt.Errorf("error al insertar datos: %v", err)
	}

	fmt.Printf("Documento insertado con DNI %s\n", response.Data.DNI)
	return nil
}

func updateExistingDocument(response *Response, collection *mongo.Collection, filter bson.M, currentTime time.Time) error {
	update := bson.M{
		"$set": bson.M{
			"nombre":           response.Data.Nombre,
			"apellido_paterno": response.Data.ApellidoPaterno,
			"apellido_materno": response.Data.ApellidoMaterno,
			"domicilio":        response.Data.Domicilio,
			"source":           response.Source,
			"updatedat":        currentTime,
		},
	}
	_, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return fmt.Errorf("error al actualizar datos: %v", err)
	}

	fmt.Printf("Documento actualizado con DNI %s\n", response.Data.DNI)
	return nil
}
