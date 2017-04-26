package dao

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func check(err error) bool {
	return err == nil
}

// Create creates an object in the collection.
// Returns a bool based on whether or not function executed correctly.
func Create(collection *mgo.Collection, object bson.M) bool {
	err := collection.Insert(object)
	return check(err)
}

// ReadOne reads one item that matches the query and writes it to the target struct.
// Returns a bool based on whether or not function executed correctly.
func ReadOne(collection *mgo.Collection, query bson.M, target interface{}) bool {
	err := collection.Find(query).One(target)
	return check(err)
}

// ReadAll reads all items matching the query and writes them all to the target struct.
// Returns a bool based on whether or not function executed correctly.
func ReadAll(collection *mgo.Collection, query bson.M, target []struct{}) bool {
	err := collection.Find(query).All(target)
	return check(err)
}

// UpdateOne updates one item that matches the query to the updated bson
// Returns a bool based on whether or not function executed correctly.
func UpdateOne(collection *mgo.Collection, query bson.M, update bson.M) bool {
	err := collection.Update(query, update)
	return check(err)
}

// UpdateAll updates all item matching the query to the updated bson
// Returns a bool based on whether or not function executed correctly.
func UpdateAll(collection *mgo.Collection, query bson.M, update bson.M) bool {
	changeInfo, err := collection.UpdateAll(query, update)
	fmt.Println(changeInfo)
	return check(err)
}

// DeleteOne finds a matching document and removes it from the database
// Returns a bool based on whether or not function executed correctly.
func DeleteOne(collection *mgo.Collection, query bson.M) bool {
	err := collection.Remove(query)
	return check(err)
}

// DeleteAll finds all matching documents and removes all from database
// Returns a bool based on whether or not function executed correctly.
func DeleteAll(collection *mgo.Collection, query bson.M) bool {
	changeInfo, err := collection.RemoveAll(query)
	fmt.Println(changeInfo)
	return check(err)
}
