package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var service = s3.New(session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")})))

func main() {
	listBuckets()
	// listFiles("first1321431")
	// uploadFile("first1321431", "secondTest.txt", "text/plain")
	downloadFile("first1321431", "secondTest", "txt")
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

func uploadFile(bucket string, file string, mediaType string) {
	//make sure to provide full filename. EX test.txt
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err.Error())
	}
	bytesToSend := bytes.NewReader(dat)
	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(file),
		ACL:           aws.String("public-read"),
		Body:          bytesToSend,
		ContentLength: aws.Int64(int64(len(dat))),
		ContentType:   aws.String(mediaType),
		Metadata: map[string]*string{
			"Key": aws.String("MetadataValue"),
		},
	}
	service.PutObject(params)
}

func downloadFile(bucket string, file string, mediaType string) error {
	//just add filename only, don't need .
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file + "." + mediaType),
	}
	result, err := service.GetObject(params)
	if err != nil {
		fmt.Println(err.Error())
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	fileErr := ioutil.WriteFile(file+"_new."+mediaType, buf.Bytes(), 0644)
	if fileErr != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("file write success")
	return fileErr
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
