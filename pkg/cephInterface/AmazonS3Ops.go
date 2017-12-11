package cephInterface

// AmazonS3Ops implements s3 get, put and delete using the Amazon client library.
// Initially the Amazon library was too buggy, but Marcus Watt of the ceph
// team debugged it for me. I expect most people will use the Amazon library.

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Proto satisfies operation by doing rest operations.
type S3Proto struct {
	Prefix   string
	S3Bucket string
	S3Key    string
	S3Secret string
	Verbose  bool
}

var svc *s3.S3
var awsLogLevel = aws.LogOff

// Get does a get operation from an s3Protocol target and times it,
func (p S3Proto) Get(path string, oldRc string) error {
	defer T.Begin(p.Prefix, path)()

	file, err := ioutil.TempFile("/tmp", "loadTesting")
	if err != nil {
		log.Fatalf("Unable to create a temp file,  %v", err)
	}
	defer os.Remove(file.Name()) // nolint

	downloader := s3manager.NewDownloaderWithClient(svc)
	initial := time.Now() //              				***** Response time starts
	numBytes, err := downloader.Download(file,
		// file is an op.Writer At, see aws.WriteAtBuffer, whcih is a []byte
		// https://gist.github.com/jboelter/ecfb08d6a18440ac16d93b5183aad207
		//buff := &aws.WriteAtBuffer{}
		//s3dl := s3manager.NewDownloader(awsSession)
		//n, err := s3dl.Download(buff, &s3.GetObjectInput{
		//	Bucket: aws.String(bucket),
		//	Key:    aws.String(key),
		//})
	    // 	if err != nil {
		//	fmt.Fprintln(os.Stderr, err)
		//	os.Exit(1)
		//}
		//n2, err := io.Copy(os.Stdout, bytes.NewReader(buff.Bytes()))
		&s3.GetObjectInput{
			Bucket: aws.String(p.S3Bucket),
			Key:    aws.String(path),
		})
	responseTime := time.Since(initial) // 				***** Response time ends
	if err != nil {
		rc := errorCodeToHTTPCode(err)
		fmt.Printf("%s %f 0 0 %d %s %d GET\n",
			initial.Format("2006-01-02 15:04:05.000"),
			responseTime.Seconds(), numBytes, path, rc)

		// Extract and reportPerformance the failure, iff possible
		return nil
	}
	fmt.Printf("%s %f 0 0 %d %s 200 GET\n",
		initial.Format("2006-01-02 15:04:05.000"),
		responseTime.Seconds(), numBytes, path)
	return nil
}

// Put puts a file and times it
// error return is used only by mkLoadTestFiles  FIXME
func (p S3Proto) Put(path, size, oldRC string) error {
	defer T.Begin(path, size)
	return fmt.Errorf("put is not implemented yet")
	//if conf.Debug {
	//	log.Printf("in AmazonS3Put(%s, %s, %d)\n", p.prefix, path, size)
	//}
	//
	//file, err := os.Open(junkDataFile)
	//if err != nil {
	//	return fmt.Errorf("Unable to open junk-data file %s, %v", junkDataFile, err)
	//}
	//defer file.Close() // nolint
	//lr := io.LimitReader(file, size)
	//
	//if svc == nil {
	//	return fmt.Errorf("missing service %v", svc)
	//}
	//uploader := s3manager.NewUploaderWithClient(svc)
	//initial := time.Now() //              				***** Response time starts
	//_, err = uploader.Upload(&s3manager.UploadInput{
	//	Bucket: aws.String(conf.S3Bucket),
	//	Key:    aws.String(path),
	//	Body:   lr,
	//})
	//responseTime := time.Since(initial) // 				***** Response time ends
	//// FIXME swap this around
	//if err == nil {
	//	fmt.Printf("%s %f 0 0 %d %s 201 PUT\n",
	//		initial.Format("2006-01-02 15:04:05.000"),
	//		responseTime.Seconds(), size, path)
	//	alive <- true
	//	return nil
	//}
	//// This doesn't seem to do what one exoects: FIXME?
	//// reqerr, ok := err.(awserr.RequestFailure)
	////if ok {
	////	log.Printf("%s %f 0 0 %d %s %d GET\n",
	////		initial.Format("2006-01-02 15:04:05.000"),
	////		responseTime.Seconds(), size, path, reqerr.StatusCode)
	////	alive <- true
	//// return nil
	////}
	//fmt.Printf("%s %f 0 0 %d %s 4XX GET\n",
	//	initial.Format("2006-01-02 15:04:05.000"),
	//	responseTime.Seconds(), size, path)
	//alive <- true
	//return fmt.Errorf("unable to upload %q to %q, %v", path, conf.S3Bucket, err)
}

// mustCreateService creates a connection to an s3-compatible server.
func  (p S3Proto) mustCreateService(myEndpoint string, awsLogLevel aws.LogLevelType) *s3.S3 {
	defer T.Begin(myEndpoint, awsLogLevel)()

	if p.S3Key == "" {
		log.Fatal("called mustCreateService with no s3 params, internal error\n")
	}
	if p.Verbose {
		awsLogLevel = aws.LogDebugWithSigning | aws.LogDebugWithHTTPBody |
			aws.LogDebugWithRequestErrors
	}
	token := ""
	creds := credentials.NewStaticCredentials(p.S3Key, p.S3Secret, token)
	_, err := creds.Get()
	if err != nil {
		log.Fatalf("bad credentials: %s\n", err)
	}
	cfg := aws.NewConfig().
		WithLogLevel(awsLogLevel).
		WithRegion("canada").
		WithEndpoint(myEndpoint).
		WithDisableSSL(true).
		WithS3ForcePathStyle(true).
		WithCredentials(creds)
	sess, err := session.NewSession() // There is a session.Must() for convenience
	if err != nil {
		log.Fatalf("bad session=%v\n", err)
	}
	svc = s3.New(sess, cfg)
	return svc
}

// Init makes sure we have an amazon s3 session and any other prerequisites.
func (p S3Proto) Init() {
	if svc == nil {
		svc = p.mustCreateService(p.Prefix, awsLogLevel)
	}
}

// errorCodeToHTTPCode is wimpey!
// only a few codes (eg, 404) are implemented
func errorCodeToHTTPCode(err error) int {
	aerr, ok := err.(awserr.Error)
	if !ok {
		return -2 // not from aws
	}
	reqErr, ok := aerr.(awserr.RequestFailure)
	if !ok {
		return -1 // not a request failure
	}
	// A service error occurred, it has an HTTP code
	return reqErr.StatusCode()
}
