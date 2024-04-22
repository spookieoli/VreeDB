package Server

import (
	"VreeDB/AccessDataHUB"
	"VreeDB/ApiKeyHandler"
	"VreeDB/Logger"
	"VreeDB/Utils"
	"VreeDB/Vdb"
	"VreeDB/Vector"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"
)

// RouteProvider is the global variable that contains all the routes
var RouteProvider *Routes

// NewRoutes returns a new Routes struct
func NewRoutes(db *Vdb.Vdb) *Routes {
	return &Routes{templates: template.Must(template.ParseGlob("templates/*.gohtml")), DB: db,
		ApiKeyHandler: ApiKeyHandler.ApiHandler, SessionKeys: make(map[string]time.Time), AData: AccessDataHUB.AccessList.ReadChan}
}

// ValidateCookie validates cookies
func (r *Routes) validateCookie(req *http.Request) bool {

	// If empty - all access is granted
	if len(ApiKeyHandler.ApiHandler.ApiKeyHashes) == 0 {
		return true
	}

	// Get the cookie
	cookie, err := req.Cookie("VreeDB")
	if err != nil {
		return false
	}

	// check if the cookie is in the map
	if _, ok := r.SessionKeys[cookie.Value]; ok {
		return true
	}
	return false
}

// CreateCookie creates a session cookie with a sessionkey (uuid)
func (r *Routes) createCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "VreeDB",
		Value:    Utils.Utils.CreateUUID(),
		HttpOnly: true,
		Secure:   true,
	}
	http.SetCookie(w, cookie)
	r.SessionKeys[cookie.Value] = time.Now()
}

// DeleteCookie is called on logout
func (r *Routes) deleteCookie(w http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie("VreeDB")
	if err != nil {
		return
	}
	delete(r.SessionKeys, cookie.Value)
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}

func (r *Routes) renderTemplate(templateName string, w http.ResponseWriter, data any) error {
	// Parse the template file anew every time
	// This pattern is not performance-efficient for production but is useful for development
	t := r.templates

	if os.Getenv("ENV") == "DEV" {
		t = template.Must(template.ParseFiles(fmt.Sprintf("templates/%s.gohtml", templateName)))
	}

	//  Render the template
	err := t.ExecuteTemplate(w, fmt.Sprintf("%s.gohtml", templateName), data)

	return err
}

/* ROUTES */

// Login is the login page
func (r *Routes) Login(w http.ResponseWriter, req *http.Request) {
	r.AData <- "LOGIN"
	if req.Method == "GET" && req.URL.Path == "/login" {
		err := r.renderTemplate("login", w, nil)
		if err != nil {
			panic(err.Error())
		}
	} else if req.Method == "POST" && req.URL.Path == "/login" {
		// Get the POST data
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing form"))
			return
		}
		// Check if the ApiKey is valid
		if r.ApiKeyHandler.CheckApiKey(req.FormValue("password")) {
			r.createCookie(w)
			http.Redirect(w, req, "/", http.StatusSeeOther)
			return
		}

		// Send User back to login
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return

	} else {
		// Send the user to the login page
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}
}

// Logout is the route to delete the cookie and so logout the user
func (r *Routes) Logout(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" && req.URL.Path == "/logout" {
		r.deleteCookie(w, req)
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}
}

// Index page
func (r *Routes) Index(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" && req.URL.Path == "/" {
		// Check if there are ApiKeys in the system
		if len(r.ApiKeyHandler.ApiKeyHashes) == 0 || r.validateCookie(req) {
			err := r.renderTemplate("index", w, NewData())
			if err != nil {
				panic(err.Error())
			}
		} else {
			// Redirect to the login Page
			http.Redirect(w, req, "/login", http.StatusSeeOther)
			return
		}
	}
}

// Delete deletes a Collection // TODO Rename to DeleteCollection
func (r *Routes) Delete(w http.ResponseWriter, req *http.Request) {
	r.AData <- "TRANSACTION"
	if req.Method == http.MethodDelete && strings.ToLower(req.URL.String()) == "/delete" {
		// Limit the size of the request
		req.Body = http.MaxBytesReader(w, req.Body, 5000)
		// Parse the form
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing form"))
			return
		}
		// load the request into the DeleteCollection via json decode
		dc := &DeleteCollection{}
		err = json.NewDecoder(req.Body).Decode(dc)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error decoding json"))
			return
		}

		// Check if the ApiKey is valid
		if r.validateCookie(req) || r.ApiKeyHandler.CheckApiKey(dc.ApiKey) { // added cookiecheck - because button ui
			// Delete all Classifiers of the Collection
			r.DB.Collections[dc.Name].DeleteAllClassifiers()

			// Call the function in the Vdb
			err = r.DB.DeleteCollection(dc.Name)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			// Send the success or error message to the client
			w.WriteHeader(http.StatusOK)
			status := struct {
				Status  string `json:"status"`
				Message string `json:"message"`
			}{
				Status:  "success",
				Message: "Collection deleted",
			}
			json.NewEncoder(w).Encode(status)
			return
		}
		// Send the unauthorized message to the client
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}
}

// We have some static js / css files without showing the filelist
func static(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CreateCollection creates a new Collection
func (r *Routes) CreateCollection(w http.ResponseWriter, req *http.Request) {
	r.AData <- "TRANSACTION"
	if req.Method == http.MethodPost && strings.ToLower(req.URL.String()) == "/createcollection" {
		// Limit the size of the request
		req.Body = http.MaxBytesReader(w, req.Body, 5000)
		// Parse the form
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing form"))
			return
		}
		// load the request into the CollectionCreator via json decode
		cc := &CollectionCreator{}
		err = json.NewDecoder(req.Body).Decode(cc)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error decoding json"))
			return
		}

		// Check if Auth is valid
		if r.ApiKeyHandler.CheckApiKey(cc.ApiKey) || r.validateCookie(req) {

			// Check if name is empty
			if cc.Name == "" || cc.Dimensions == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Variables Missing"))
				return
			}

			// Check if Collection exists
			if _, ok := r.DB.Collections[cc.Name]; ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Collection with name " + cc.Name + " allready exists"))
				return
			}

			// There is a wait bool - if true the function will wait for the collection to be created
			if cc.Wait {
				// Choose distance function from Distancefunction string
				if strings.ToLower(cc.DistanceFunction) != "euclid" {
					cc.DistanceFunction = "cosine"
				}
				err = r.DB.AddCollection(cc.Name, cc.Dimensions, cc.DistanceFunction)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				// Send the success or error message to the client
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Collection created"))
				return
			} else {
				// Create the Collection
				go r.DB.AddCollection(cc.Name, cc.Dimensions, cc.DistanceFunction)
				// Send the success or error message to the client
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Collection created"))
				return
			}
		}

		// Not authorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	// Notice the user that the route is not found under given information
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
	return
}

// ListCollections lists all the Collections
func (r *Routes) ListCollections(w http.ResponseWriter, req *http.Request) {
	r.AData <- "TRANSACTION"
	if req.Method == http.MethodGet && strings.ToLower(req.URL.String()) == "/listcollections" {
		// Create CollectionList type
		cl := &CollectionList{}

		// Check if Auth is valid
		if r.ApiKeyHandler.CheckApiKey(cl.ApiKey) || r.validateCookie(req) {
			// Get the Collections
			collections := r.DB.ListCollections()
			cl.Collections = collections

			// Send the collections to the client
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(cl)
			return
		}

		// Not authorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}
	// Notice the user that the route is not found under given information
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
	return
}

// AddPoint adds a point to a Collection
func (r *Routes) AddPoint(w http.ResponseWriter, req *http.Request) {
	r.AData <- "ADD"
	if req.Method == http.MethodPut && strings.ToLower(req.URL.String()) == "/addpoint" {
		// Limit the size of the request
		req.Body = http.MaxBytesReader(w, req.Body, 10000000)
		// Parse the form
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing form"))
			return
		}
		// load the request into the Point via json decode
		p := &Point{}
		err = json.NewDecoder(req.Body).Decode(p)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error decoding json"))
			return
		}

		// Check if Auth is valid
		if r.ApiKeyHandler.CheckApiKey(p.ApiKey) || r.validateCookie(req) {

			// Checks if the CollectionName and the Vector are set
			if p.CollectionName == "" || p.Vector == nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Missing required fields"))
				return
			}

			// Check if Collection exists
			if _, ok := r.DB.Collections[p.CollectionName]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Collection does not exist"))
				return
			}

			// Add the point to the Collection
			v := Vector.NewVector(p.Id, p.Vector, &p.Payload, p.CollectionName)
			err = r.DB.Collections[p.CollectionName].Insert(v)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			// Send the success or error message to the client
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Point added"))
			return
		}

		// Not authorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return

	}

	// Notice the user that the route is not found under given information
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
	return
}

// AddPointBatch adds a batch of points to a Collection
func (r *Routes) AddPointBatch(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPut && strings.ToLower(req.URL.String()) == "/addpointbatch" {
		// Parse the form
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing form"))
			return
		}

		// load the request into the PointBatch via json decode
		pb := PointBatch{}
		err = json.NewDecoder(req.Body).Decode(&pb)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error decoding json"))
			return
		}

		// Check if Auth is valid
		if r.ApiKeyHandler.CheckApiKey(pb.ApiKey) || r.validateCookie(req) {
			// Name, Vector are required
			if pb.CollectionName == "" || pb.Points == nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Missing required fields"))
				return
			}

			// Check if Collection exists
			if _, ok := r.DB.Collections[pb.CollectionName]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Collection does not exist"))
				return
			}

			// Add the points to the Collection
			go func() {
				for _, p := range pb.Points {
					r.AData <- "ADD"
					d := p.Payload // This is no longer necessary from GO >= 1.22
					v := Vector.NewVector(p.Id, p.Vector, &d, pb.CollectionName)
					err = r.DB.Collections[pb.CollectionName].Insert(v)
					if err != nil {
						Logger.Log.Log("Error in BulkAdd: " + err.Error())
						return
					}
				}
			}()

			// Send the success or error message to the client
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Points bulk added"))
			return
		}

		// Not authorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return

	}

	// Notice the user that the route is not found under given information
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
	return
}

// DeletePoint deletes a point from a Collection
func (r *Routes) DeletePoint(w http.ResponseWriter, req *http.Request) {
	r.AData <- "DELETE"
	if req.Method == http.MethodDelete && strings.ToLower(req.URL.String()) == "/deletepoint" {
		// Limit the size of the request
		req.Body = http.MaxBytesReader(w, req.Body, 5000)
		// Parse the form
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing form"))
			return
		}
		// load the request into the DeletePoint via json decode
		dp := &DeletePoint{}
		err = json.NewDecoder(req.Body).Decode(dp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error decoding json"))
			return
		}

		// Check if Auth is valid
		if r.ApiKeyHandler.CheckApiKey(dp.ApiKey) || r.validateCookie(req) {

			// Check if Collection exists
			if _, ok := r.DB.Collections[dp.CollectionName]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Collection does not exist"))
				return
			}

			// Delete the point from the Collection
			err = r.DB.Collections[dp.CollectionName].Delete(dp.Id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		}

		// not authorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return

		// Send the success or error message to the client
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Point deleted"))
		return
	}

	// Notice the user that the route is not found under given information
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
	return
}

// Search searches for the nearest neighbours of the given target vector
func (r *Routes) Search(w http.ResponseWriter, req *http.Request) {
	r.AData <- "SEARCH"
	if req.Method == http.MethodGet {
		// Parse the form
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing form"))
			return
		}
		// load the request into the Point via json decode
		p := &Point{}
		err = json.NewDecoder(req.Body).Decode(p)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err.Error())
			w.Write([]byte("Error decoding json"))
			return
		}

		// Check if Auth is valid
		if r.ApiKeyHandler.CheckApiKey(p.ApiKey) || r.validateCookie(req) {

			// Check if possible Filter is valid
			if err := p.ValidateFilter(); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			// Name, Vector are required
			if p.CollectionName == "" || p.Vector == nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Missing required fields"))
				return
			}

			// Check if Collection exists
			if _, ok := r.DB.Collections[p.CollectionName]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Collection does not exist"))
				return
			}

			// Search for the nearest neighbours
			var queue *Utils.HeapControl
			if p.Depth == 0 {
				queue = Utils.NewHeapControl(3)
			} else {
				queue = Utils.NewHeapControl(p.Depth)
			}

			// Set the resultset
			var results []*Utils.ResultSet

			// Check if Index is set
			switch p.Index {
			case nil:
				results = r.DB.Search(p.CollectionName, Vector.NewVector(p.Id, p.Vector, &p.Payload, ""), queue,
					p.MaxDistancePercent, p.Filter)
			default:
				results = r.DB.IndexSearch(p.CollectionName, Vector.NewVector(p.Id, p.Vector, &p.Payload, ""),
					queue, p.MaxDistancePercent, p.Filter, p.Index.IndexName, p.Index.IndexValue)
			}

			// Send the results to the client
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(results)
			return
		}

		// not authorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return

	}

	// Notice the user that the route is not found under given information
	fmt.Println(req.Method, req.URL.String())
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
	return
}

// TrainClassifier trains a classifier
func (r *Routes) TrainClassifier(w http.ResponseWriter, req *http.Request) {
	r.AData <- "SYSTEMEVENT"
	if req.Method == http.MethodPut && strings.ToLower(req.URL.String()) == "/trainclassifier" {
		// Parse the form
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing form"))
			return
		}
		// load the request into the TrainClassifier via json decode
		tc := &Classifier{}
		err = json.NewDecoder(req.Body).Decode(tc)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error decoding json"))
			return
		}

		// Check if Auth is valid
		if r.ApiKeyHandler.CheckApiKey(tc.ApiKey) || r.validateCookie(req) {

			// Check if Collection exists
			if _, ok := r.DB.Collections[tc.CollectionName]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Collection does not exist"))
				return
			}

			// Check if Collection is ClassifierReady
			if !r.DB.Collections[tc.CollectionName].ClassifierReady {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Collection is not ready for classification"))
				return
			}

			// Check if Type exists
			if tc.Type == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Type is missing"))
				return
			}

			// Create the classifier in the collection
			err = r.DB.Collections[tc.CollectionName].AddClassifier(tc.ClassifierName, tc.Type, tc.Loss)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			// Train the classifier non blocking
			go func() {
				err := r.DB.Collections[tc.CollectionName].TrainClassifier(tc.ClassifierName, tc.Degree, tc.C, tc.Epochs)
				if err != nil {
					Logger.Log.Log(err.Error())
				}
			}()

			// Send the success or error message to the client
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Classifier created and training started"))
			return
		}

		// not authorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return

	}

	// Notice the user that the route is not found under given information
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
	return
}

// DeleteClassifier will delete a classifier
func (r *Routes) DeleteClassifier(w http.ResponseWriter, req *http.Request) {
	r.AData <- "TRANSACTION"
	if req.Method == http.MethodDelete && strings.ToLower(req.URL.String()) == "/deleteclassifier" {
		// Limit the size of the request
		req.Body = http.MaxBytesReader(w, req.Body, 5000)

		// Parse the form
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing form"))
			return
		}

		// load the request into the DeleteClassifier via json decode
		dc := &DeleteClassifier{}
		err = json.NewDecoder(req.Body).Decode(dc)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error decoding json"))
			return
		}

		// Check if Auth is valid
		if r.ApiKeyHandler.CheckApiKey(dc.ApiKey) || r.validateCookie(req) {

			// Check if Collection exists
			if _, ok := r.DB.Collections[dc.CollectionName]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Collection does not exist"))
				return
			}

			// Delete the classifier from the collection
			r.DB.Collections[dc.CollectionName].DeleteClassifier(dc.ClassifierName)

			// Log the deletion
			Logger.Log.Log("Classifier " + dc.ClassifierName + " in Collection " + dc.CollectionName + " deleted")

			// Send the success or error message to the client
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Classifier deleted"))
			return
		}

		// Not authorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return

	}
	// Notice the user that the route is not found under given information
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
	return
}

// Classify will classify a vector
func (r *Routes) Classify(w http.ResponseWriter, req *http.Request) {
	r.AData <- "CLASSIFY"
	if req.Method == http.MethodGet && strings.ToLower(req.URL.String()) == "/classify" {
		// Parse the form
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing form"))
			return
		}

		// load the request into the Classify via json decode
		c := &Classify{}
		err = json.NewDecoder(req.Body).Decode(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error decoding json"))
			return
		}

		// Check if Auth is valid
		if r.ApiKeyHandler.CheckApiKey(c.ApiKey) || r.validateCookie(req) {

			// Check if Collection exists
			if _, ok := r.DB.Collections[c.CollectionName]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Collection does not exist"))
				return
			}

			// Check if Classifier exists
			if _, ok := r.DB.Collections[c.CollectionName].Classifiers[c.ClassifierName]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Classifier does not exist"))
				return
			}

			// Check if the vector is of the right dimension
			if len(c.Vector) != r.DB.Collections[c.CollectionName].VectorDimension {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Vector has wrong dimension it should be " + fmt.Sprint(r.DB.Collections[c.CollectionName].VectorDimension) + " but is " + fmt.Sprint(len(c.Vector)) + " long"))
				return
			}

			// Classify the vector
			class := r.DB.Collections[c.CollectionName].Classifiers[c.ClassifierName].Predict(c.Vector)

			// Send the class to the client
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// Type switch - a classifier can have various returns
			switch class.(type) {
			case int:
				json.NewEncoder(w).Encode(struct {
					Class int `json:"class"`
				}{
					Class: class.(int),
				})
			case []float64:
				json.NewEncoder(w).Encode(struct {
					Class []float64 `json:"class"`
				}{
					Class: class.([]float64),
				})
			case float64:
				json.NewEncoder(w).Encode(struct {
					Class float64 `json:"class"`
				}{
					Class: class.(float64),
				})
			}
			return
		}

		// Not authorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return

	}
	// Notice the user that the route is not found under given information
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
	return
}

// CreateApiKey creates a new ApiKey
func (r *Routes) CreateApiKey(w http.ResponseWriter, req *http.Request) {
	r.AData <- "SYSTEMEVENT"
	if req.Method == http.MethodPut && strings.ToLower(req.URL.String()) == "/createapikey" {
		// Parse the form
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing form"))
			return
		}

		// load the request into the ApiKeyCreator via json decode
		ac := &ApiKeyCreator{}
		err = json.NewDecoder(req.Body).Decode(ac)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error decoding json"))
			return
		}

		// Check if Auth is valid
		if r.ApiKeyHandler.CheckApiKey(ac.ApiKey) || r.validateCookie(req) {

			// Create the ApiKey
			key, err := r.ApiKeyHandler.CreateApiKey()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			// Return the apikey to the client
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ApiKeyCreator{
				ApiKey: key,
			})
			return
		}

		// not authorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}
	// Notice the user that the route is not found under given information
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
	return
}

// DeleteApiKey deletes an ApiKey
func (r *Routes) DeleteApiKey(w http.ResponseWriter, req *http.Request) {
	r.AData <- "SYSTEMEVENT"
	if req.Method == http.MethodDelete && strings.ToLower(req.URL.String()) == "/deleteapikey" {
		// Limit the size of the request
		req.Body = http.MaxBytesReader(w, req.Body, 5000)

		// Parse the form
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing form"))
			return
		}

		// load the request into the DeleteApiKey via json decode
		da := &ApiKeyCreator{}
		err = json.NewDecoder(req.Body).Decode(da)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error decoding json"))
			return
		}

		// Check if Auth is valid
		if r.ApiKeyHandler.CheckApiKey(da.ApiKey) || r.validateCookie(req) {
			// Delete the ApiKey
			err = r.ApiKeyHandler.DeleteApiKey(da.ApiKey)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			// Send the success or error message to the client
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ApiKey deleted"))
			return
		}
		// Not authorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}
	// Notice the user that the route is not found under given information
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
	return
}

// CreateIndex will create an index
func (r *Routes) CreateIndex(w http.ResponseWriter, req *http.Request) {
	r.AData <- "SYSTEMEVENT"
	if req.Method == http.MethodPut && strings.ToLower(req.URL.String()) == "/createindex" {
		// Parse the form
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing form"))
			return
		}

		// load the request into the IndexCreator via json decode
		ic := &IndexCreator{}
		err = json.NewDecoder(req.Body).Decode(ic)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error decoding json"))
			return
		}

		// Check if all field of the IndexCreator are set
		if ic.ApiKey == "" || ic.CollectionName == "" || ic.IndexName == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing required fields"))
			return
		}

		// Check if Auth is valid
		if r.ApiKeyHandler.CheckApiKey(ic.ApiKey) || r.validateCookie(req) {
			// Create the Index
			err = r.DB.Collections[ic.CollectionName].CreateIndex(ic.IndexName, ic.IndexName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			// Send the success or error message to the client
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Index created"))
			return
		}

		// not authorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}
	// Notice the user that the route is not found under given information
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
	return

}

// GetAccessData will return the AccessData
func (r *Routes) GetAccessData(w http.ResponseWriter, req *http.Request) {
	r.AData <- "SYSTEMEVENT"
	if req.Method == http.MethodPost && strings.ToLower(req.URL.String()) == "/getaccessdata" {
		// Check if Auth is valid
		if r.validateCookie(req) {
			// Get the data
			accessList := AccessDataHUB.AccessList.GetData()
			// Send the AccessData to the client
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(accessList)
			return
		}

		// Not authorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}
	// Notice the user that the route is not found under given information
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
	return
}

// showapikey will show the apikey
func (r *Routes) ShowApiKey(w http.ResponseWriter, req *http.Request) {
	r.AData <- "SYSTEMEVENT"
	if req.Method == http.MethodGet && strings.ToLower(req.URL.String()) == "/showapikey" {
		// This will only work if there is no APIKEY
		if len(ApiKeyHandler.ApiHandler.ApiKeyHashes) == 0 {
			// Create the APIKEY
			key, err := r.ApiKeyHandler.CreateApiKey()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			// Create the data for the template to show the apikey
			data := struct {
				Data string
			}{
				Data: key,
			}
			//  Show the template
			err = r.renderTemplate("showapikey", w, data)
			if err != nil {
				Logger.Log.Log(err.Error())
			}
			return
		} else {
			// Redirect to login
			http.Redirect(w, req, "/login", http.StatusSeeOther)
			return
		}
	}
}
