package models

import (
	"gopkg.in/mgo.v2/bson"
)

// User is a collection
// ID is a bson object id
// Username is a string representing the username they use to log in
// Password is a string representing the password they use to log in TODO: bcrypt
type User struct {
	ID         bson.ObjectId `bson:"_id,omitempty"`
	Username   string
	Password   string
	CustomerID string `bson:", omitempty"`
}
