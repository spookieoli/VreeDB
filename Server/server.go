package Server

import (
	"VreeDB/ArgsParser"
	"VreeDB/Boot"
	"VreeDB/Logger"
	"VreeDB/Vdb"
	"context"
	"fmt"
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
func NewServer(ip string, port int, certfile string, keyfile string, secure bool) *Server {
	// Create the Server Object - booting up the DB
	server := &Server{Ip: ip, Port: port, DB: Vdb.DB, CertFile: certfile, KeyFile: keyfile, Secure: secure}

	// Define a new ServeMux
	mux := http.NewServeMux()

	// Start the Webserver
	server.Server = &http.Server{
		Addr:              server.Ip + ":" + strconv.Itoa(server.Port),
		Handler:           mux,
		ReadHeaderTimeout: time.Second * 600,
		WriteTimeout:      time.Second * 600,
		IdleTimeout:       time.Second * 600,
	}

	// Start  the bootup
	server.DB.Collections = Boot.NewBootUp().Boot()

	// Add the routes
	server.addRoutes(mux)
	return server
}

// addRoutes adds all routes to the server
func (s *Server) addRoutes(mux *http.ServeMux) {
	// Get all the Routes out of the Routeprovider
	routes := NewRoutes(s.DB)
	v := reflect.ValueOf(routes)
	for i := 0; i < v.NumMethod(); i++ {
		// get the Name of the Route
		name := v.Type().Method(i).Name
		// Get the Route
		route := v.MethodByName(name).Interface().(func(http.ResponseWriter, *http.Request))
		if name == "Index" {
			mux.HandleFunc("/", route)
			continue
		}
		mux.HandleFunc("/"+strings.ToLower(name), route)
	}
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", static(fileServer)))
}

// Start starts the server
func (s *Server) Start() {
	Logger.Log.Log("Server is listening on "+s.Ip+":"+strconv.Itoa(s.Port), "INFO")
	if s.Secure {
		s.Server.ListenAndServeTLS(s.CertFile, s.KeyFile)
	} else {
		s.Server.ListenAndServe()
	}
}

// Shutdown shuts down the server
func (s *Server) Shutdown() {
	Logger.Log.Log("Server is shutting down", "INFO")

	// Create a context with timeout - we give all Clients 10 seconds to finish their requests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	err := s.Server.Shutdown(ctx)
	if err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	fmt.Println("Server gracefully stopped")
}
