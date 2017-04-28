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
	"github.com/icrowley/fake"
	"github.com/julienschmidt/httprouter"
)

var service = s3.New(session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")})))

var sess *mgo.Session

//BucketRequest : struct for S3 buckets request
type BucketRequest struct {
	Username   string
	Bucketname string
}

//File : struct to define img files
type File struct {
	DisplayName string
	URL         string
	MediaType   string
	Width       int64
	Height      int64
	BucketName  string
}

//User : struct to define users
type User struct {
	Username string
	Password string
	Buckets  []models.Bucket
}

//Getters Start
func (u User) listUserFiles() {
	for i := 0; i < len(u.Buckets); i++ {
		// listFiles(u.Buckets[i])
	}
}

func (f File) getDimensionString() string {
	return string(f.Width) + "x" + string(f.Height)
}

//Getters End

func makeFile(filename string, bucket string, size1 int64, size2 int64) *File {
	name := filename[0 : len(filename)-4]
	mediaType := filename[len(filename)-3:]
	url := "https://s3.amazonaws.com/" + bucket + "/" + filename
	return &File{DisplayName: name, URL: url, MediaType: mediaType, Width: size1, Height: size2}
}

func makeBucket(name string) *models.Bucket {
	first := fake.Word()
	nums := fake.DigitsN(8)
	randName := first + nums
	// fmt.Println("randName:", randName)
	return &models.Bucket{DisplayName: name, BucketName: randName}
}
func createBucket(bucket string) {
	params := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}
	service.CreateBucket(params)
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

func listFiles(bucket string) []string {
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}
	resp, err := service.ListObjectsV2(params)
	listArray := make([]string, 0)
	check(err)
	for _, key := range resp.Contents {
		listArray = append(listArray, *key.Key)
	}
	return listArray
}

//Upload a file
func uploadFile(bucket string, file string, mediaType string) {
	dat, err := ioutil.ReadFile(file + "." + mediaType)
	check(err)
	bytesToSend := bytes.NewReader(dat)
	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(file + "." + mediaType),
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

//Download a file
func downloadFile(bucket string, file string, mediaType string) error {
	fmt.Println("file:", file)
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file + "." + mediaType),
	}
	result, err := service.GetObject(params)
	check(err)
	fmt.Println("bucket:", bucket)
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	fileErr := ioutil.WriteFile(file+"_new."+mediaType, buf.Bytes(), 0644)
	if fileErr != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("file write success")
	return fileErr
}

//Delete Buckets
func deleteBucket(bucket models.Bucket) {
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

func handleLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-type", "application/json")
	c := sess.DB("dropbox-clone").C("user")
	request := bsonify(r)
	doc := models.User{}
	fmt.Println(request)
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
		return
	}
	createFailureResponse("something went wrong", &w)
	return
}

func handleBucket(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := sess.DB("dropbox-clone").C("user")
	body := BucketRequest{}
	check(json.NewDecoder(r.Body).Decode(&body))
	b := makeBucket(ps.ByName("bucketname"))
	query := bson.M{"username": body.Username}
	user := models.User{}
	readSuccess := dao.ReadOne(c, query, &user)
	if readSuccess {
		createBucket(b.BucketName)
		user.Buckets = append(user.Buckets, *b)
		dao.UpdateOne(c, query, bson.M{"$set": bson.M{"buckets": user.Buckets}})
		fmt.Println("user", user.Username)
		createResponse(user, &w)
		return
	}
	createFailureResponse("Something went wrong!", &w)
}

func handleFileUpload(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	body := File{}
	check(json.NewDecoder(r.Body).Decode(&body))
	f := makeFile(body.DisplayName+"."+body.MediaType, body.BucketName, body.Height, body.Width)
	uploadFile(body.BucketName, f.DisplayName, f.MediaType)
	fmt.Println("file created!")
	createResponse(bson.M{}, &w)
}

func handleFileDownload(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := sess.DB("dropbox-clone").C("user")
	user := models.User{}
	query := bson.M{"username": "dev"}
	readSuccess := dao.ReadOne(c, query, &user)
	if readSuccess {
		downloadFile(ps.ByName("bucketname"), ps.ByName("filename"), ps.ByName("mediatype"))
		fmt.Println("download worked!!")
		createResponse(user, &w)
	}
}

func handleListFiles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := sess.DB("dropbox-clone").C("user")
	query := bson.M{"username": "dev"}
	user := models.User{}
	readSuccess := dao.ReadOne(c, query, &user)
	if readSuccess {
		list := listFiles(ps.ByName("bucketname"))
		createResponse(list, &w)
		return
	}
	createFailureResponse("Something went wrong", &w)
}

func main() {
	var connErr error
	sess, connErr = mgo.Dial("mongodb://admin:drowssap@ds051980.mlab.com:51980/dropbox-clone")
	check(connErr)

	defer sess.Close()
	sess.SetMode(mgo.Monotonic, true)

	router := httprouter.New()
	//ROUTES
	router.POST("/api/authenticate", handleLogin)
	router.POST("/api/register", handleRegister)
	router.POST("/api/bucket/:bucketname", handleBucket)
	router.POST("/api/file/upload", handleFileUpload)
	router.GET("/api/file/download/:bucketname/:filename/:mediatype", handleFileDownload)
	router.GET("/api/user/files/:bucketname", handleListFiles)

	http.ListenAndServe(":8080", router)
}
