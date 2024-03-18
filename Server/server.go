package Server

import (
	"VectoriaDB/Boot"
	"VectoriaDB/Logger"
	"VectoriaDB/Vdb"
	"VectoriaDB/Vector"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	Ip     string
	Port   int
	Server *http.Server
	DB     *Vdb.Vdb
}

// NewServer returns a new Server
func NewServer(ip string, port int) *Server {
	// Create the Server Object - booting up the DB
	server := &Server{Ip: ip, Port: port, DB: Vdb.DB}
	// Start the Webserver
	server.Server = &http.Server{
		Addr:              server.Ip + ":" + strconv.Itoa(server.Port),
		Handler:           nil,
		ReadHeaderTimeout: time.Second * 60,
		WriteTimeout:      time.Second * 15,
		IdleTimeout:       time.Second * 60,
	}
	// Start  the bootup
	server.DB.Collections = Boot.NewBootUp().Boot()

	// Create a test Collection
	err := server.DB.AddCollection("test", 800, "euclid")
	if err != nil {
		panic(err)
	}

	// Now we are adding 5000000 Vectors to the Collection with random values, 800 dimensions
	for i := 0; i < 1000; i++ {
		data := make([]float64, 800)
		for j := 0; j < 800; j++ {
			data[j] = rand.Float64()
		}
		payload := make(map[string]interface{})
		err := server.DB.Collections["test"].Insert(Vector.NewVector("", data, &payload, "test"))
		if err != nil {
			panic(err)
		}
	}

	// Add one special vector to check if we find the right one
	data := make([]float64, 800)
	for j := 0; j < 800; j++ {
		data[j] = 0.5
	}
	payload := make(map[string]interface{})
	err = server.DB.Collections["test"].Insert(Vector.NewVector("YEAH", data, &payload, "test"))
	if err != nil {
		panic(err)
	}

	for i := 0; i < 1000; i++ {
		data := make([]float64, 800)
		for j := 0; j < 800; j++ {
			data[j] = rand.Float64()
		}
		payload := make(map[string]interface{})
		err := server.DB.Collections["test"].Insert(Vector.NewVector("", data, &payload, "test"))
		if err != nil {
			panic(err)
		}
	}

	// Add the routes
	server.addRoutes()
	return server
}

// addRoutes adds all routes to the server
func (s *Server) addRoutes() {
	// Get all the Routes out of the Routeprovider
	routes := NewRoutes(s.DB)
	v := reflect.ValueOf(routes)
	for i := 0; i < v.NumMethod(); i++ {
		// get the Name of the Route
		name := v.Type().Method(i).Name
		// Get the Route
		route := v.MethodByName(name).Interface().(func(http.ResponseWriter, *http.Request))
		if name == "Index" {
			http.HandleFunc("/", route)
			continue
		}
		http.HandleFunc("/"+strings.ToLower(name), route)
	}
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", static(fileServer)))
}

// Start starts the server
func (s *Server) Start() {
	Logger.Log.Log("Server is listening on " + s.Ip + ":" + strconv.Itoa(s.Port))
	log.Fatal(s.Server.ListenAndServe())
}
