package Server

import (
	"VectoriaDB/ArgsParser"
	"VectoriaDB/Boot"
	"VectoriaDB/Logger"
	"VectoriaDB/Vdb"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	Ip         string
	Port       int
	Server     *http.Server
	DB         *Vdb.Vdb
	ArgsParser *ArgsParser.ArgsParser
	CertFile   string
	KeyFile    string
	Secure     bool
}

// NewServer returns a new Server
func NewServer(ip string, port int, certfile, keyfile string, secure bool) *Server {
	// Create the Server Object - booting up the DB
	server := &Server{Ip: ip, Port: port, DB: Vdb.DB, CertFile: certfile, KeyFile: keyfile, Secure: secure}

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
	if s.Secure {
		log.Fatal(s.Server.ListenAndServeTLS(s.CertFile, s.KeyFile))
	} else {
		log.Fatal(s.Server.ListenAndServe())
	}
}
