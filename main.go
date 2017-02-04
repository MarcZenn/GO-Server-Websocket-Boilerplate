/*Instruct compiler to make this file an executable
* using 'package main.
* Keyword main tells us this package will be compiled to a program executable
* instead of .a file which is a static library archive file.
*
 */
package main

import (
	"log"
	"net/http"

	r "gopkg.in/gorethink/gorethink.v3"
)

/*Entry point for the executable which in this case is our GO Server.
 * This is a good place to connect to our database as well.
 * We will store that connection as 'session'.
 *
 */
func main() {
	// Connect to our DB. Client will need to access this connection aka session so its been added to Client struct in client.go
	session, err := r.Connect(r.ConnectOpts{
		Address:  "localhost:28015",
		Database: "whatsitlike",
	})

	if err != nil {
		log.Panic(err.Error())
	}

	router := NewRouter(session)              // return a properly initialized Router Struct instance.
	router.Handle("subscribe", subscribeFeed) // subscribe to feed upon site visit.
	router.Handle("new upload", addUpload)    // tell router when receive 'new upload' message call add upload function and handle DB insert.
	// Define new user by handler in handlers.go
	// Define Filtering by handler in handlers.go
	// Define New question handler in handlers.go
	// Define Upvote method handler in handlers.go
	// Define Downvote method handler in handlers.go

	http.Handle("/", router)          // pass in our router.go handler as the handler
	http.ListenAndServe(":4000", nil) // Listen & Port
}
