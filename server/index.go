package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var service = s3.New(session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")})))

func main() {
	listBuckets()
	listFiles("first1321431")
}

func listFiles(bucket string) *s3.ListObjectsOutput {
	params := &s3.ListObjectsInput{
		Bucket: aws.String("bucket"),
	}
	resp, err := service.ListObjects(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
	}

	// Pretty-print the response data.
	fmt.Println(resp)
	return resp
}

//ListBucketsRequest : GET buckets?
func listBuckets() *s3.ListBucketsOutput {
	result, err := service.ListBuckets(nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Buckets:")
	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}
	return result
}

//Make Buckets
//Not Work atm
// func createBuckets(bucket string) {
// 	resp := service.CreateBucket(&s3.CreateBucketInput{
// 		Bucket: aws.String(bucket)})
// }

//Error handler
func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
