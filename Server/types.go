package Server

import (
	"VreeDB/ApiKeyHandler"
	"VreeDB/Filter"
	"VreeDB/NN"
	"VreeDB/Utils"
	"VreeDB/Vdb"
	"VreeDB/Vector"
	"html/template"
	"time"
)

// CollectionCreator is the struct that creates a Collection in the VDB, when send by REST
type CollectionCreator struct {
	ApiKey           string `json:"api_key"` // Must not be present in the request
	Name             string `json:"name"`
	DistanceFunction string `json:"distance_function"`
	Dimensions       int    `json:"dimensions"`
	Wait             bool   `json:"wait"`
}

// Used to delete a Collection, when send by REST
type DeleteCollection struct {
	ApiKey string `json:"api_key"`
	Name   string `json:"name"`
}

// CollectionList is the struct that lists all the Collections (NAMES) in the VDB, when send by REST
type CollectionList struct {
	ApiKey      string   `json:"api_key"`
	Collections []string `json:"collections"`
}

type IndexName struct {
	IndexName  string `json:"index_name"`
	IndexValue any    `json:"index_value"`
}

// Point is the struct that adds a point to a Collection, when send by REST
type Point struct {
	Id                 string                 `json:"id"` // Must not be present in the request
	ApiKey             string                 `json:"api_key"`
	CollectionName     string                 `json:"collection_name"`
	Vector             []float64              `json:"vector"`
	Payload            map[string]interface{} `json:"payload"`              // Optional
	Depth              int                    `json:"depth"`                // Must not be present in the request default 3
	Wait               bool                   `json:"wait"`                 // Must not be present in the request default false
	MaxDistancePercent float64                `json:"max_distance_percent"` // Must not be present in the request default 0.0 (no limit)
	Index              *IndexName             `json:"index"`                // Must not be present in the request default ""
	Filter             *[]Filter.Filter       `json:"filter"`               // Must not be present in the request default nil
	GetVectors         bool                   `json:"get_vectors"`          // Must not be present in the request default false
	GetId              bool                   `json:"get_id"`               // Must not be present in the request default false
}

type PointItem struct {
	Id      string                 `json:"id"` // Must not be present in the request
	Vector  []float64              `json:"vector"`
	Payload map[string]interface{} `json:"payload"` // Optional
}

// PointBatch is the struct that adds a batch of points to a Collection, when send by REST
type PointBatch struct {
	ApiKey         string      `json:"api_key"`
	CollectionName string      `json:"collection_name"`
	Points         []PointItem `json:"points"`
}

// Result is a struct that contains the result of a search
type Result struct {
	Vector   *Vector.Vector `json:"vector"`
	Distance float64        `json:"distance"`
}

// SearchResult is the struct that contains the result of a search
type SearchResult struct {
	Results []*Result `json:"results"`
}

// Routes is the struct that contains all the routes
type Routes struct {
	templates     *template.Template
	DB            *Vdb.Vdb
	ApiKeyHandler *ApiKeyHandler.ApiKeyHandler
	SessionKeys   map[string]time.Time
	AData         chan string
}

// Collection will display Collection related stuff
type Collection struct {
	Name            string   `json:"name"`
	NodeCount       int      `json:"node_count"`
	DistanceFunc    string   `json:"distance_func"`
	DiagonalLength  float64  `json:"diagonal_length"`
	Classifier      []string `json:"classifier"`
	ClassifierReady bool     `json:"classifier_ready"`
	Dimensions      int      `json:"dimensions"`
}

// RuntimeData is the struct that will be used to display Application runtime data
type RuntimeData struct {
	CollectionCount int
	RamUsage        float64
	FreeRam         float64
	Percent         float64
	Uptime          int64
	ApiKeyExists    bool
}

// Classifier is the struct that will be used to create a new classifier
type Classifier struct {
	ApiKey         string          `json:"api_key"`
	CollectionName string          `json:"collection_name"`
	ClassifierName string          `json:"classifier_name"`
	Degree         int             `json:"degree"`
	C              float64         `json:"c"`
	Epochs         int             `json:"epochs"`
	Type           string          `json:"type"`
	Loss           string          `json:"loss"`
	Batchsize      int             `json:"batchsize"`
	Architecture   *[]NN.LayerJSON `json:"architecture"`
}

// DeleteClassifier is the struct that will be used to delete a classifier, when send by REST
type DeleteClassifier struct {
	ApiKey         string `json:"api_key"`
	CollectionName string `json:"collection_name"`
	ClassifierName string `json:"classifier_name"`
}

// Classify will be the struct that will be used to classify a vector, when send by REST
type Classify struct {
	ApiKey         string    `json:"api_key"`
	CollectionName string    `json:"collection_name"`
	ClassifierName string    `json:"classifier_name"`
	Vector         []float64 `json:"vector"`
}

// Data will be the struct that will be used to display the web page
type Data struct {
	Collections []Collection
	Application RuntimeData
}

// DeletePoint is the struct that will be used to delete a point from a Collection, when send by REST
type DeletePoint struct {
	ApiKey         string           `json:"api_key"`
	CollectionName string           `json:"collection_name"`
	Id             []string         `json:"id"`
	Filter         *[]Filter.Filter `json:"filter"` // Must not be present in the request default nil
}

type Cartesian struct {
	X, Y, Z float64
}

type Geo struct {
	ApiKey    string
	Latitude  float64
	Longitude float64
	Altitude  float64
}

// ApiKeyCreator is the struct that will be used to create a new Api key
type ApiKeyCreator struct {
	ApiKey string `json:"api_key"`
}

// ShowTrainProgress shows the progress of the training of the neural network
type ShowTrainProgress struct {
	CollectionName string `json:"collection_name"`
	ClassifierName string `json:"classifier_name"`
}

type IndexCreator struct {
	ApiKey         string `json:"api_key"`
	CollectionName string `json:"collection_name"`
	IndexName      string `json:"index_name"`
}

type TSNE struct {
	ApiKey         string  `json:"api_key"`
	CollectionName string  `json:"collection_name"`
	Dimensions     int     `json:"dimensions"`
	Learningrate   float64 `json:"learning_rate"`
	Iterations     int     `json:"iterations"`
}

type TSNEData struct {
	CollectionName string      `json:"collection_name"`
	Vector         [][]float64 `json:"vector"`
}

// ValidateFilter will validate the filters in Point
func (p *Point) ValidateFilter() error {
	if p.Filter != nil {
		for _, filter := range *p.Filter {
			if err := filter.Op.IsValid(); err != nil {
				return err
			}
		}
	}
	return nil
}

// NewData creates new Data Structure for the web page
func NewData() Data {
	data := Data{}
	// Add all the Collections
	for _, collection := range Vdb.DB.Collections {
		data.Collections = append(data.Collections, Collection{Name: collection.Name, NodeCount: len(*collection.Space),
			DistanceFunc: collection.DistanceFuncName, DiagonalLength: collection.DiagonalLength,
			Classifier: collection.ClassifierToSlice(), ClassifierReady: collection.ClassifierReady,
			Dimensions: collection.VectorDimension})
	}
	data.Application = RuntimeData{RamUsage: Utils.Utils.GetMemoryUsage(), FreeRam: Utils.Utils.GetAvailableRAM(),
		Uptime: 0, Percent: (Utils.Utils.GetMemoryUsage() / Utils.Utils.GetAvailableRAM()) * 100,
		CollectionCount: len(Vdb.DB.Collections), ApiKeyExists: !ApiKeyHandler.ApiHandler.CheckIfEmpty()}

	return data
}
