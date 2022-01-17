package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/justinas/alice"
)

// net/http provides HTTP client & server implementations

/*
 *
 * the client
 *
 */

// The client makes HTTP requests, and receives HTTP responses

// There is a default client instance included, providing an empty &http.Client{}
// but you should always use your own, as it has no timeout
var dontUseDefault = http.DefaultClient

// You only need to create one single http.Client for your entire program, as it properly handles
// multiple simultaneous requests
func newClient() *http.Client {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return client
}

// When you want to make a request, you create a new *http.Request instance with the
// http.NewRequestWithContext function
func newReq() *http.Request {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://jsonplaceholder.typicode.com/todos/1", nil)
	if err != nil {
		panic(err)
	}
	return req
}

// Once you have an *http.Request instance, you can set any headers via the
// Headers firled of the instance. Once you're done, you can use the Do()
// method on the http.Client with your http.Request which returns an
// http.Response
func makeReq() *http.Response {
	c := newClient()
	req := newReq()
	req.Header.Add("X-My-Client", "Learning Go")
	resp, err := c.Do(req)
	if err != nil {
		panic(err)
	}

	return resp
}

// The response has several fields with information on the request:
func handleResponse(r *http.Response) {
	code := r.StatusCode // The numeric code of the response status
	codeText := r.Status // The text response of the status, e.g. "200 OK"
	headers := r.Header  // Response headers

	contentType := headers.Get("Content-Type") // headers is map[string][]string

	fmt.Printf("%d\n%s\n%s", code, codeText, contentType)
}

// Response bodies can be used json.Decoder to process REST API responses
func parseJson(r *http.Response) {
	// the response body is an io.ReadCloser, which means it can be used to parse json
	body := r.Body

	var data struct {
		UserID   int    `json:"userId"`
		ID       int    `json:"id"`
		Title    string `json:"title"`
		Complete bool   `json:"completed"`
	}

	err := json.NewDecoder(body).Decode(&data)
	if err != nil {
		panic(err)
	}
}

/*
 *
 * HTTP Server
 *
 */

// The HTTP server is built around the concept of an http.Server and the http.Handler interface.
// http.Server listens for HTTP requests, requests to the server are handled by implementations of
// http.Handler, which have a single method ServeHTTP
/*
https://pkg.go.dev/net/http?utm_source=gopls#Handler
 	type Handler interface {
		ServeHTTP(ResponseWriter, *Request)
	}
*/

// ServeHTTP takes two arguments - we've already looked at http.Request
// lets take a look at http.Response Writer

/*
https://pkg.go.dev/net/http?utm_source=gopls#ResponseWriter
	type ResponseWriter interface {
		Header() Header
		Write([]byte) (int, error)
		WriteHeader(statusCode int)
	}
*/

// Lets define our own handler
type NewHandler struct{}

func (n NewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// the methods in http.ResponseWriter must be called in a certain order
	// the first, Header(), gives you an instnace of http.Header so you can set
	// any headers you need.
	// If you don't need it, don't call it

	// Next, call WriteHeader() with the HTTP status code for your response
	w.WriteHeader(http.StatusAccepted) // if you are sending a 200, you can skip it this

	// Write() sets the body for the response

	w.Write([]byte("Hello, world!\n"))
}

// Now that we have our server, we can make our handler
func newServer() http.Server {
	s := http.Server{
		Addr:         ":8000",           // TCP address to listen. host:port - if not provided, listen on all hosts on port 80
		ReadTimeout:  30 * time.Second,  // Time to wait to read request headers
		WriteTimeout: 30 * time.Second,  // Time to wait for the write of the response
		IdleTimeout:  120 * time.Second, // Time to wait for the next request when keep-alives are enabled
		Handler:      NewHandler{},      // We invoke our handler when we get a request to our server
	}

	return s
}

// The ListenAndServeMethod starts the the HTTP server
func listen(s http.Server) {
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

// The main problem with our server is that it only handles one path,
// thankfully *http.ServeMux meets the http.Handler interface, and
// serves as a router - sending the requests to the correct
// http.Handler instance

// You can look at the details of how http.ServeMux does this here:
// https://cs.opensource.google/go/go/+/refs/tags/go1.17.6:src/net/http/server.go;l=2416 ServeHTTP
// https://cs.opensource.google/go/go/+/refs/tags/go1.17.6:src/net/http/server.go;l=2361 Handler
// but, essentially, http.ServeMux implements ServeHTTP(w http.ResponseWriter, r *http.Request),
// satisfying the http.Handler interface, so it can be passed in as Handler to http.Server
// It's ServeHTTP implementation calls the Handler, which parses the path and returns the
// http.Handler used for the request, and then calls that handlers ServeHTTP method

// Servemux instances are the most common way of implementing multiple handlers

// That sounded a bit confusing, so let's test this out
func newServeMux() {
	m := http.NewServeMux() // creates a blank *http.ServeMux

	// create two handlers, an alive and ready path
	m.HandleFunc("/alive", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok!\n"))
	})
	m.HandleFunc("ready", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("always\n"))
	})

	s := http.Server{
		Addr:         ":8000",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      m, // Our handler now handles two requests, /alive and /ready
	}

	s.ListenAndServe()
}

// http.ServeMux instances, since they themselves route requests to http.Handler instances,
// and since http.ServeMux is itself an instance of a http.Handler interface, an http.ServeMux
// can handle other http.ServeMux instances.
// This lets us route nested paths
func parentChildMux() {
	user := http.NewServeMux()
	user.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a user!"))
	})

	record := http.NewServeMux()
	record.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a record!"))
	})

	mux := http.NewServeMux()

	// mux will handle both the user and record path.
	// http.StripPrefix removes the part of the path that's already been processed,
	// as the previous handles don't expect the first path
	mux.Handle("/user/", http.StripPrefix("/user", user))
	mux.Handle("/record", http.StripPrefix("/record", record))
}

// One last pitfall

// The http library offers functions that work with the package instance of
// *http.ServeMux, declared as:
var dontUseDefaultMux = http.DefaultServeMux

// These functions are http.ListenAndServe, http.HandleFunc, and http.ListenAndServeTLS

// http.ListenAndServe and http.ListenAndServeTLS both serve with the default http server,
// which as mentione does not contain properties such as timeouts.
func dontUselistenAndServeDefault() {
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServeTLS(":80", "", "", nil)
}

// Using these methods can allow third party packages that have added
// additional handlers to http.DefaultServeMux to inject vulnerabilities

/*
 *
 * middleware
 *
 */

// There is no middleware type in go, just a pattern involving http.Handler instances:

// RequestTimer is a pice of middleware, taking in a http.Handler record and returning one
// that times the request
func RequestTimer(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		end := time.Now()
		log.Printf("request time for %ss: %v", r.URL.Path, end.Sub(start))
	})
}

// TerribleSecurityProvider is a piece of middleware which
// returns a function that returns a handler
// The middleware itself is checking the header for the correct password, and if so,
// serving the handler that was pased to it
var securityMsg = []byte("You didn't give the secret password\n")

func TerribleSecurityProvider(password string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Secret-Password") != password {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write(securityMsg)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

// Now for the implementation!
// our mux handles the /hello path with TerribleSecurityProvider,
// instantiated as ts.
// As ts is itself a function that takes a handler, we can then pass in
// RequestTimer to it.
//
// If ts passes its security check, it runs h.ServeHTTP, in this case RequestTimer.
// If not, it returns, ending the request and breaking the chain.
//
// If it passes, it calls RequestTimer which itself takes a handler that writes "Hello, World"
// RequestTimer starts timing the request, and then runs h.ServeHTTP, ultimately serving "Hello, world\n"
// on path /hello
func muxWithMiddleware() {
	ts := TerribleSecurityProvider("PASSWORD")
	mux := http.NewServeMux()

	mux.Handle("/hello", ts(RequestTimer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello, world!\n"))
		}))))
}

/*
 *
 * Postscript
 *
 */

// Third party packages:

// Alice:
// https://justinas.org/alice-painless-middleware-chaining-for-go
// Makes chaining middleware as simple as:

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello, world!"))
}

func timeoutHandler(h http.Handler) http.Handler {
	return http.TimeoutHandler(h, 1*time.Second, "timed out")
}

func aliceWithMiddleWare() {
	ts := TerribleSecurityProvider("PASSWORD")
	handler := http.HandlerFunc(helloWorldHandler)

	a := alice.New(ts, timeoutHandler).Then(handler)

	s := http.Server{
		Addr:         ":8000",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      a,
	}

	s.ListenAndServe()
}

// Gorilla:
// A bit too big to cover here but in a nutshell it has a
// mux package:
// https://github.com/gorilla/mux
//
// which allows easy use of creating dynamic paths such as
// /user/{user_id}/name
