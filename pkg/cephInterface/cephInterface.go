package cephInterface
// package cephInterface accesses ceph, currently via authenticated
// s3-protocol get, put and delete using the Amazon client library.
//
// Initially the Amazon library was too buggy, but Marcus Watt of the
// ceph team debugged it for me.
//
// I expect most people will use the Amazon library, even though there
// is a native RADOS one for Go, which looks significantly better...

import (
	"fmt"
	"time"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/davecb/trace"
	"sync"
	"strconv"
	"log"
	"encoding/json"
)


// S3Proto satisfies operation by doing rest operations.
type S3Proto struct {
	endpoint    string
	s3Key       string
	s3Secret    string
	verbose     bool
	svc         *s3.S3
	awsLogLevel aws.LogLevelType
	            trace.Trace
}

var singletonS3 *S3Proto
var once sync.Once
var logger		*log.Logger

// New creates a single s3 interface
// stretch goal -- do this with a pipe
func New(t trace.Trace, x *log.Logger) *S3Proto {
	if t == nil {
		 t = trace.New(nil, true)
	}
	logger = x
	defer t.Begin()()
	var p = S3Proto{
		//endpoint: "http://10.121.10.201:7480",   // IAD3, also 202, ...
		endpoint: "http://10.92.10.201:7480",  // AMS1
		verbose:  	false,
		s3Key:    	"91V7FH4MNMXQW2WRBAZI",
		s3Secret: 	"bhZIl6LPMKjm0dHW5zyb23OwNXWsJxAdVLIms5Xh",
		Trace: 		t,
	}
	once.Do(func() {
		mustCreateService(&p)
		singletonS3 = &p
	})
	return singletonS3
}

// Get does a head-and-get operation from an s3Protocol target and times it.
// the time will have an extra half-RTT in it because of the s3 architecture
func (p S3Proto) Get(key, bucket string) ([]byte, map[string]string, int, error) {
	var rc int
	var head = make(map[string]string)
	defer p.Begin(p.endpoint, key, bucket)()

	if p.svc == nil { // FIXME, belt and suspenders, drop
		panic(fmt.Errorf("in cephInterface.Get, p.svc = %v", p.svc))
	}

	// get head, see if object exists
	initial := time.Now() //                        ***** Time starts
	latency, head, rc, err := getHead(p, bucket, key, initial, head)
	if err != nil {
		reportPerformance(initial, latency, 0.0,
			0.0, 0, key, rc,	"GET")
		return nil, head, rc, fmt.Errorf(
			"failed in svc.headObject, %v", err)
	}
	if rc != 200 {
		reportPerformance(initial, latency, 0.0,
			0.0, 0, key, rc,	"GET")
		return nil, head, rc, nil
	}

	// get body (only) of object
	xferTime, buff, numBytes, rc, err := getBody(p, bucket, key)
	if err != nil {
		reportPerformance(initial, latency, xferTime,
			0.0, numBytes, key, rc,	"GET")
		return nil, head, rc, fmt.Errorf(
			"failed in downloader.Download, %v", err)
	}
	if numBytes > 0 {
		head["Content-Length"] = strconv.FormatInt(numBytes, 10)
	}
	if rc != 200 {
		reportPerformance(initial, latency, xferTime, 0.0, numBytes, key, rc,	"GET")
		return nil, head, rc, nil
	}
	reportPerformance(initial, latency, xferTime,
		0.0, numBytes, key, rc,	"GET")
	return buff.Bytes(), head, rc, nil
}

// Put puts a file and times it
func (p S3Proto) Put(contents []byte, path, bucket string) error {
	defer p.Begin("<contents>", path, bucket)()

	// if contents is empty, fail and then log.
	// Caller may ignore the failure
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

// Head does a head operation on an s3Protocol target and times it.
func (p S3Proto) Head(key, bucket string) (map[string]string, int, error) {
	var rc int
	var head= make(map[string]string)
	defer p.Begin(key, bucket)()

	if p.svc == nil { // FIXME, belt and suspenders, drop
		panic(fmt.Errorf("in cephInterface.Get, p.svc = %v", p.svc))
	}

	// get head, see if object exists
	initial := time.Now() //                        ***** Time starts
	latency, head, rc, err := getHead(p, bucket, key, initial, head)
	if err != nil {
		reportPerformance(initial, latency, 0.0,
			0.0, 0, key, rc, "HEAD")
		return head, rc, fmt.Errorf(
			"failed in svc.headObject, %v", err)
	}
	reportPerformance(initial, latency, 0.0,
		0.0, 0, key, rc, "HEAD")
	return head, rc, nil
}


// getHead -- get the head information, specifically including headers
// See also https://docs.aws.amazon.com/goto/WebAPI/s3-2006-03-01/HeadObjectOutput
func getHead(p S3Proto, bucket string, key string, initial time.Time, headers map[string]string) (time.Duration,
	map[string]string, int, error) {
	var rc = 200

	s3head, err := p.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	latency := time.Since(initial) // 	        	***** Latency ends
	if err != nil {
		rc = errorCodeToHTTPCode(err)
		if rc < 0 {
			// it's a real error, say so
			logger.Printf("INFO, HeadObject failed %v", err)
			return latency, nil, rc, err
		}
		// normal case: just a non-success code from server
		return latency, nil, rc, nil
	}
	// CAVEAT, this only does part of the implemented subset
	setHeader(headers, "Accept-Ranges", s3head.AcceptRanges)
	setHeader(headers,"Content-Disposition",  s3head.ContentDisposition)
	setHeader(headers,"Content-Encoding",  s3head.ContentEncoding)
	setHeader(headers,"Content-Type", s3head.ContentType)
	setHeader(headers,"Content-Language",  s3head.ContentLanguage)
	s := strconv.FormatInt(*s3head.ContentLength, 10)
	setHeader(headers,"Content-Length", &s)
	if s3head.DeleteMarker != nil {
		headers["x-amz-delete-marker"] = strconv.FormatBool(*s3head.DeleteMarker)
	}
	setHeader(headers,"ETag", s3head.ETag)
	setHeader(headers,"x-amz-expiration",  s3head.Expiration)
	setHeader(headers,"Expires", s3head.Expires)
	if s3head.LastModified != nil {
		headers["Last-Modified"] = s3head.LastModified.Format(time.RFC850)
	}
	if s3head.Metadata != nil {
		j, _ := json.Marshal(s3head.Metadata)
		headers["Metadata"] = string(j)
	}
	if s3head.PartsCount != nil {
		headers["x-amz-mp-parts-count"] = strconv.FormatInt(*s3head.PartsCount, 10)
	}
	setHeader(headers,"x-amz-replication-status", s3head.ReplicationStatus)
	setHeader(headers,"x-amz-storage-class",  s3head.StorageClass)
	setHeader(headers,"x-amz-version-id",  s3head.VersionId)

	return latency, headers, rc, err
}
func setHeader(headers map[string]string, k string, v *string) {
	if v != nil {
		headers[k] = *v
	}
}

// minions -- these do work and disambiguate err from "rc != 200"

// getBody -- get a body
func getBody(p S3Proto, bucket string, key string) (time.Duration,
	*aws.WriteAtBuffer, int64, int, error) {

	initial := time.Now() //             ***** Time starts
	buff := &aws.WriteAtBuffer{}
	downloader := s3manager.NewDownloaderWithClient(p.svc)
	numBytes, err := downloader.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	xferTime := time.Since(initial)  //  ***** Time ends
	if err != nil {
		rc := errorCodeToHTTPCode(err)
		if rc < 0 {
			// an error, not a 404 or the like
			logger.Printf("INFO, downloader.Download failed, %v", err)
			return xferTime, buff, numBytes, rc, err
		}
		return xferTime, buff, numBytes, rc, nil
	}
	return xferTime, buff, numBytes, 200, nil
}

// mustCreateService creates a connection to an s3-compatible server.
func mustCreateService(p *S3Proto) {
	defer p.Begin(p)()

	if p.s3Key == "" {
		panic(fmt.Errorf("called mustCreateService with no s3 params, internal error"))
	}
	if p.verbose {
		p.awsLogLevel = aws.LogDebugWithSigning | aws.LogDebugWithHTTPBody |
			aws.LogDebugWithRequestErrors
	}
	token := ""
	creds := credentials.NewStaticCredentials(p.s3Key, p.s3Secret, token)
	_, err := creds.Get()
	if err != nil {
		panic(fmt.Errorf("in mustCreateService, credentials.NewStaticCredentials() = %v", err))
	}
	cfg := aws.NewConfig().
		WithLogLevel(p.awsLogLevel).
		WithRegion("canada").
		WithEndpoint(p.endpoint).
		WithDisableSSL(true).
		WithS3ForcePathStyle(true).
		WithCredentials(creds)
	sess, err := session.NewSession()
	if err != nil {
		panic(fmt.Errorf("in mustCreateService, session.NewSession() = %v", err))
	} else {
		p.Printf("in mustCreateService, session.NewSession() = %v\n", sess)
	}
	p.svc = s3.New(sess, cfg)
	if p.svc == nil {
		panic(fmt.Errorf("in mustCreateService, s3.New returned a nil session"))
	}  else {
		p.Printf("in mustCreateService, s3.New() = %v\n", p.svc)
	}
}

// errorCodeToHTTPCode is wimpey! only a few codes (eg, 404) are implemented
// s3 is pretty hackey in places...
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

// reportPerformance in standard format
func reportPerformance(initial time.Time, latency, xferTime,
	thinkTime time.Duration, length int64, key string, rc int,
	op string) {

	fmt.Printf("%s %f %f %f %d %s %d %s\n",
		initial.Format("2006-01-02 15:04:05.000"),
		latency.Seconds(), xferTime.Seconds(), thinkTime.Seconds(),
		length, key, rc, op)
}
