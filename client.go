/*Responsibilities of the Client/Struct:
 *
 * Send & Receive messages to Browser. Client will need access to websocket for this.
 *
 */
package main

import (
	"github.com/gorilla/websocket"
	r "gopkg.in/gorethink/gorethink.v3"
)

// Message struct is exported and used elsewhere. Mainly for sending back to client.
type Message struct {
	Name string      `json:"name"` // changes key from Name to name
	Data interface{} `json:"data"` // changes key from Data to data
}

// FindHandler type of function is exported and used elsewhere. Basically tells our http Handle function which handler function to use depending on the event.
type FindHandler func(string) (Handler, bool)

// Client struct is used to
type Client struct {
	send         chan Message
	socket       *websocket.Conn // type pointer to socket connection
	findHandler  FindHandler
	session      *r.Session // add our DB session aka connection so its available to our handlers
	stopChannels map[int]chan bool
}

// NewStopChannel GO Channel
func (client *Client) NewStopChannel(stopkey int) chan bool {
	stop := make(chan bool)
	client.stopChannels[stopkey] = stop
	return stop
}

// Method we use to read messages from the client over the websocket. Notice we are using a receiver in parantheses (client *Client) which makes this a method of the Client struct and can access other Client struct properties like 'client.socket'.
func (client *Client) Read() {
	var message Message
	for {
		if err := client.socket.ReadJSON(&message); err != nil {
			break
		}
		// Call findHandler to know which handler to call. If handler in router map value matches key then call it.
		if handler, found := client.findHandler(message.Name); found {
			handler(client, message.Data)
		}
	}
	// close connection once finished.
	client.socket.Close()
}

// Method we use to send messages to the client over the websocket. Notice we are using a receiver in parantheses (client *Client) which makes this a method of the Client struct and can access other Client struct properties like 'client.socket'.
func (client *Client) Write() {
	for msg := range client.send {
		if err := client.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	// close connection once finished.
	client.socket.Close()
}

// Close and shit
func (client *Client) Close() {
	for _, ch := range client.stopChannels {
		ch <- true
	}
	close(client.send)
}

// NewClient is a function to create a new Client struct instance (by using its location in memory). Create another GO channel of type Message within that instance. This will be used by router.go ServeHTTP function to communicate with browser and send back messages.
func NewClient(socket *websocket.Conn, findHandler FindHandler, session *r.Session) *Client { // pass pointer to socket conn.
	return &Client{ // return pointer for newly instanstiated Client object
		send:         make(chan Message),
		socket:       socket,
		findHandler:  findHandler,
		session:      session, // initialize session to passed session value
		stopChannels: make(map[int]chan bool),
	}
}
