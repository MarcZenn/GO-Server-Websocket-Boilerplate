/*Custom GO Router:
 *
 * Router is essentialy a route handler. It will handle initial HTTP request along with data events. A common way to handle these http requests is via GO's http package func Handle() function. Its first param will be a string like normal but its 2nd param will be a unique type of Struct aka a handler interface. Basically the Struct we provide must implement the handler interface. Basically you can give structs behavioral methods as needed by defining them inside interfaces.
 */
package main

// Import packages
import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	r "gopkg.in/gorethink/gorethink.v3"
)

/*
 * Websocket Upgrader Function from GO to Hijack HTTP protocol and use Sockets instead:
 * ---------------------------------------------------------------------------------------
 * What Upgrader does is attempt to hijack the connection and switch protocols from http to websocket (while enforcing same-origin policy). We'll use it in our hanlder func below to upgrade our connection. Notice that when calling Upgrader we use curly braces instead of parentheses.
 *
 *
 * Websockets in modern browsers allow connection to any host which is different than your typical http requests. Basically there is no browser-based same-origin policy, its up to the server to enforce a same-origin policy when using websockets. We must manually determine if its ok for browser to make websocket connection.
 */
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // return true indicates allow connection from any origin.
}

// Router Struct
type Router struct {
	rules   map[string]Handler // Type map is a dictionary of key/value pairs where the keys are strings and the values are our Handler functions aka methods for each type of route we need passed in as a string i.e. "new upload", "upvote", "downvote" etc.
	session *r.Session
}

// NewRouter makes a new Router struct instance once called. We need it because maps must be initialized before being used. To do this we need to create an instance like so:
func NewRouter(session *r.Session) *Router {
	return &Router{
		rules:   make(map[string]Handler),
		session: session,
	}
}

//Handler is a type function for all our handler functions. Pass what you need here for use in any Handler.
type Handler func(*Client, interface{})

// Handle here is simply what calls our Route handlers for all Router instance method handlers. ie. "new upload", "upvote", etc.
func (r *Router) Handle(msgName string, handler Handler) { // second param is handler aka addUpload in main.go. Add anything you need to its type above.

	r.rules[msgName] = handler // For key "string" its value is w/e handler function we passed in defined in main.go
}

// FindHandler function is used to determine which router "rule" aka which handler the Handle func above should use for any given inbound messages from client.
func (r *Router) FindHandler(msgName string) (Handler, bool) {
	handler, found := r.rules[msgName]
	return handler, found
}

// Our router is also responsible for handling HTTP itself. We attach this to Router struct as a method via receiver (e *Router). It will use our upgrader function to upgrade our connection from http to websockets which we will use to handle communications between server and client.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	// Call Upgrader to switch our connection to websockets. calling the Upgrader function returns more than one value as a response from the server as opposed to just one like in most languages. The returned values are (*Conn, error) which we need to store as variables. Golang will infer the type of the variable. This states: set var socket and var err to Upgrader's returned vaules (*Conn, error) respectively and infer the type using :=. This is the normal pattern for handling errors in GO. It is called a 'brief statement', and can only be used inside of function bodies not global variables.
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		// throw 500 error.
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	// This is how we'll interface with the client. We will have to communicate with the client here so makes to create a new instance of our Client struct from router.go. It will need a websocket connection to read from and write to so pass in socket to it.
	client := NewClient(socket, r.FindHandler, r.session)
	defer client.Close()
	go client.Write() // since using same socket needs to be in seperate GO routine
	client.Read()     // will use GO routine of this function call

}
