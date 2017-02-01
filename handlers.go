/*All functions in this file must of type Handler defined in router.go file. It must have the same params.
 */
package main

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	r "gopkg.in/gorethink/gorethink.v3"
)

// TODO:: Define new user by handler in handlers.go?
// TODO:: Define Filtering by handler in handlers.go
// TODO:: Define New question handler in handlers.go
// TODO:: Define Upvote method handler in handlers.go
// TODO:: Define Downvote method handler in handlers.go

// ChannelStop const using iota
const (
	ChannelStop = iota
)

/*Feedlet is a struct. A Struct is a way to declare a data type. Similar to a JS prototype object to which you can add methods and properties. The way to define a method prop is by using an 'interface' as shown below. An interface is a way to assign behaviors to a data type in GO. An empty interface does not specify any methods and thus can be said that every data type in GO implements the behavior of an empty interface. Can be used as a placeholder if you don't know what the behavior will be.
type Speaker interface {
	Speak()
}
func (u Upload) Speaker() {
	fmt.Println("I am a " + u.name + " event!")
}
*/
type Feedlet struct {
	ID   string `json:"id" gorethink:"id,omitempty"` // changes key from Id to id when json
	Name string `json:"name" gorethink:"name"`       // changes key from Name to name when json
}

/*func subscribeFeed(client *Client, data interface{}) {
	// changefeed is a blocking operation so place in seperate GO routine.
	go func() {
		cursor, err := r.Table("upload").Changes(r.ChangesOpts{IncludeInitial: true}).Run(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
			return
		}
		var change r.ChangeResponse
		for cursor.Next(&change) {
			if change.NewValue != nil && change.OldValue == nil {
				client.send <- Message{"subscribing", change.NewValue} // send it all back.
			}
		}
	}()
}*/

func addUpload(client *Client, data interface{}) {
	var upload Feedlet // upload saved to DB
	err := mapstructure.Decode(data, &upload)
	if err != nil {
		client.send <- Message{"error", err.Error()} // if can't decode send err back to client.
		return
	}
	// wrap DB insert in its own go routine since initial routine is being used already by read method.
	go func() {
		// Insert into rethinkDB.
		err = r.Table("upload").Insert(upload).Exec(client.session) // execute insert.
		if err != nil {
			client.send <- Message{"error", err.Error()} // if can't insert send err back to client.
		}
	}()
}

// Goal of this handler is to send all existing feed records to the client upon site visit and then utilize a change feed to populate new records as they are uploaded.
func subscribeFeed(client *Client, data interface{}) {
	fmt.Println("subscribed")
	stop := client.NewStopChannel(ChannelStop)
	result := make(chan r.ChangeResponse)

	// grab all of our existing uploads and display on visiting site.
	cursor, err := r.Table("upload").Changes(r.ChangesOpts{IncludeInitial: true}).Run(client.session) // changefeed to return existing channel records.
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}

	go func() {
		var change r.ChangeResponse
		for cursor.Next(&change) {
			result <- change
		}
	}()

	go func() {
		for {
			select {
			case <-stop:
				cursor.Close()
				return
			case change := <-result:
				if change.NewValue != nil && change.OldValue == nil {
					// a new upload has been added
					client.send <- Message{"subscribing", change.NewValue} // subscribing event
				} else if change.NewValue == nil && change.OldValue != nil {
					// an upload has been deleted
					client.send <- Message{"delete upload", change.NewValue}
				}
			}

		}
	}()
}
