package main

import (
	"bytes"
	"dropbox-clone/server/dao"
	"dropbox-clone/server/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/julienschmidt/httprouter"
)

var service = s3.New(session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")})))

var sess *mgo.Session

//Bucket : struct for S3 buckets
type Bucket struct {
	DisplayName string
	BucketName  string
}

//File : struct to define img files
type File struct {
	DisplayName string
	URL         string
	MediaType   string
	Width       int64
	Height      int64
}

//User : struct to define users
type User struct {
	Username string
	Password string
	Buckets  []Bucket
}

//Getters Start
func (u User) listUserFiles() {
	for i := 0; i < len(u.Buckets); i++ {
		listFiles(u.Buckets[i])
	}
}

func (f File) getDimensionString() string {
	return string(f.Width) + "x" + string(f.Height)
}

func (b Bucket) getBucketName() string {
	return b.BucketName
}

//Getters End

func listFiles(bucket Bucket) {
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket.BucketName),
	}
	resp, err := service.ListObjectsV2(params)

	check(err)

	// Pretty-print the response data.
	for _, key := range resp.Contents {
		fmt.Println(*key.Key)
	}
}

//ListBucketsRequest : GET buckets?
func listBuckets() *s3.ListBucketsOutput {
	result, err := service.ListBuckets(nil)
	check(err)

	fmt.Println("Buckets:")
	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}
	return result
}

//Upload a file
func uploadFile(bucket Bucket, file File) {
	//make sure to provide full filename. EX test.txt
	dat, err := ioutil.ReadFile(file.DisplayName)
	check(err)
	bytesToSend := bytes.NewReader(dat)
	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket.BucketName),
		Key:           aws.String(file.DisplayName),
		ACL:           aws.String("public-read"),
		Body:          bytesToSend,
		ContentLength: aws.Int64(int64(len(dat))),
		ContentType:   aws.String(file.MediaType),
		Metadata: map[string]*string{
			"Key": aws.String("MetadataValue"),
		},
	}
	service.PutObject(params)
}

//Download a file
func downloadFile(bucket Bucket, file File) error {
	//just add filename only, don't need .
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket.BucketName),
		Key:    aws.String(file.DisplayName + "." + file.MediaType),
	}
	result, err := service.GetObject(params)
	check(err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	fileErr := ioutil.WriteFile(file.DisplayName+"_new."+file.MediaType, buf.Bytes(), 0644)
	if fileErr != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("file write success")
	return fileErr
}

//Make Buckets
func createBucket(bucket Bucket) {
	params := &s3.CreateBucketInput{
		Bucket: aws.String(bucket.BucketName),
	}
	service.CreateBucket(params)
}

//Delete Buckets
func deleteBucket(bucket Bucket) {
	params := &s3.DeleteBucketInput{
		Bucket: aws.String(bucket.BucketName),
	}
	service.DeleteBucket(params)
}

//Error handler
func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

//Check function to handle errors
func check(err error) {
	if err != nil {
		panic(err)
	}
}

func createResponse(data interface{}, w *http.ResponseWriter) {
	response := bson.M{"data": data, "success": true}
	json.NewEncoder(*w).Encode(&response)
	return
}

func createFailureResponse(reason string, w *http.ResponseWriter) {
	response := bson.M{"reason": reason, "success": false}
	json.NewEncoder(*w).Encode(&response)
}

func bsonify(r *http.Request) bson.M {
	body := bson.M{}
	check(json.NewDecoder(r.Body).Decode(&body))
	return body
}

func createUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := sess.DB("dropbox-clone").C("user")
	request := bsonify(r)
	writeSuccess := dao.Create(c, request)
	if writeSuccess {
		createResponse(request, &w)
		return
	}
	createFailureResponse("Something went wrong", &w)
	return
}

func handleLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-type", "application/json")
	c := sess.DB("dropbox-clone").C("user")
	request := bsonify(r)
	doc := models.User{}
	readSuccess := dao.ReadOne(c, request, &doc)
	if readSuccess {
		if doc.Username == request["username"] && doc.Password == request["password"] {
			createResponse(doc, &w)
			return
		}
		createFailureResponse("Could not find a matching username and password combination", &w)
		return
	}
	createFailureResponse("Something went very wrong", &w)
	return
}

func handleRegister(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := sess.DB("dropbox-clone").C("user")
	request := bsonify(r)
	writeSuccess := dao.Create(c, request)
	if writeSuccess {
		createResponse(request, &w)
		fmt.Println(request)
		return
	}
	createFailureResponse("something went wrong", &w)
	return
}

func handleBucketList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	listBuckets()
}

func main() {
	var connErr error
	sess, connErr = mgo.Dial("mongodb://admin:drowssap@ds051980.mlab.com:51980/dropbox-clone")
	check(connErr)

	defer sess.Close()
	sess.SetMode(mgo.Monotonic, true)

	router := httprouter.New()

	router.POST("/api/authenticate", handleLogin)
	router.POST("/api/register", handleRegister)
	router.GET("/", handleBucketList)

	http.ListenAndServe(":8080", router)
}
