package models

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/mgo.v2/bson"
)

// User is a collection
// ID is a bson object id
// Username is a string representing the username they use to log in
// Password is a string representing the password they use to log in TODO: bcrypt
type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Username string
	Password string
	Buckets  []Bucket `bson:"buckets"`
}

//Bucket : Defines buckets
type Bucket struct {
	DisplayName string `bson:"displayName"`
	BucketName  string `bson:"bucketName"`
}

var service = s3.New(session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")})))

//Check function to handle errors
func check(err error) {
	if err != nil {
		panic(err)
	}
}

func (b Bucket) listFiles() {
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(b.BucketName),
	}
	resp, err := service.ListObjectsV2(params)

	check(err)

	// Pretty-print the response data.
	for _, key := range resp.Contents {
		fmt.Println(*key.Key)
	}
}

func (b Bucket) getBucketName() string {
	return b.BucketName
}

//Make Buckets
func (b *Bucket) createBucket() {
	params := &s3.CreateBucketInput{
		Bucket: aws.String(b.BucketName),
	}
	service.CreateBucket(params)
}
