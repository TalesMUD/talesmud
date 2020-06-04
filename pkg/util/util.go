package util

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"time"
)

//HTTPWriteImage utility to write image to response writer
func HTTPWriteImage(w http.ResponseWriter, img *image.Image) {

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, *img); err != nil {
		log.Println("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("unable to write image.")
	}
}

// Route ... contains data to declare a route
type Route struct {
	Pattern     string
	Method      string
	Name        string
	HandlerFunc http.Handler
}

// Routes ... type to hold multiple routes
type Routes []Route

//HTTPResponder serves method to respond to http calls
type HTTPResponder interface {
	JSON(w http.ResponseWriter, code int, payload interface{})
	ERROR(w http.ResponseWriter, code int, err error)
}

/*
// JSON responds to the request with the given code and payload
func (app *app) JSON(w http.ResponseWriter, code int, payload interface{}) {

	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)

}

// JSON responds to the request with the given code and payload
func (app *app) ERROR(w http.ResponseWriter, code int, err error) {
	response := []byte(err.Error())
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
}
*/
// Logger function for http calls
func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}
