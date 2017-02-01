/*Instruct compiler to make this file an executable using 'package main'.... 'package somelibname' would create a library. Keyword main tells us this package will be compiled to a program executable instead of .a file which is a static library archive file.
*
*
*
 */
package main

// import packages
import (
	"log"
	"net/http"

	r "gopkg.in/gorethink/gorethink.v3"
)

// type User struct {
// 	Id   string `gorethink:"id,omitempty"`
// 	Name string `gorethink:"name"`
// }

// Entry point for the executable which in this case is our GO Server. This is a good place to connect to our DB as well. We will call that connection 'session'.
func main() {
	// Connect to our DB. Client will need to access this connection aka session so its been added to Client struct
	session, err := r.Connect(r.ConnectOpts{
		Address:  "localhost:28015",
		Database: "whatsitlike",
	})
	if err != nil {
		log.Panic(err.Error())
	}
	router := NewRouter(session) // return a properly initialized Router Struct. The new instance.

	router.Handle("subscribe", subscribeFeed) // subscribe to feed upon site visit.

	router.Handle("new upload", addUpload) // tell router when receive 'new upload' message call add upload function and handle DB insert.

	http.Handle("/", router)          // pass in our router.go handler as the handler
	http.ListenAndServe(":4000", nil) // Listen & Port
}
