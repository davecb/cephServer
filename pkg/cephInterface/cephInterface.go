package cephInterface
// package cephInterface accesses ceph, currently via authenticated
// s3-protocol get, put and delete using the Amazon client library.
// Initially the Amazon library was too buggy, but Marcus Watt of the
// ceph team debugged it for me. I expect most people will use the
// Amazon library, even though there is a native ceph one for Go.

import (
	"fmt"
	"time"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/davecb/cephServer/pkg/trace"
	"sync"
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

// New creates a single s3 interface
func New(t trace.Trace) *S3Proto {
	if t == nil {
		 t = trace.New(nil, true)
	}
	defer t.Begin(t)()
	var p = S3Proto{
		endpoint: "http://10.92.10.201:7480", // FIXME This seems to be haproxy->RCDN, as it returns fids
		//endpoint: "http://10.92.100.1:7480",  // as does this.
		//endpoint: "http://10.92.100.1:1080", connection refused
		//endpoint: "http://10.92.100.1:5666", connection reset by peer
		//endpoint: "http://10.92.100.1:6789", malformed HTTP status code
		//endpoint: "http://10.92.100.1:8443", malformed HTTP status code

		//endpoint: "http://10.92.10.201:8500", // FIXME 404 suggests it may be something
		//endpoint: "http://10.92.10.201:80",  gets ServiceStack message in html  'Endpoint' should not be empty.''
		//endpoint: "http://10.92.10.201:81", ditto

		verbose:  true,
		s3Key:    "91V7FH4MNMXQW2WRBAZI",
		s3Secret: "bhZIl6LPMKjm0dHW5zyb23OwNXWsJxAdVLIms5Xh",
		Trace: t,
	}
	once.Do(func() {
		mustCreateService(&p)
		singletonS3 = &p
	})
	return singletonS3
}

// Get does a head-and-get operation from an s3Protocol target and times SOMETHING.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/s3-2006-03-01/HeadObjectOutput
func (p S3Proto) Get(key, bucket string) ([]byte, *s3.HeadObjectOutput, error) {
	defer p.Begin(p.endpoint, key, bucket)()

	if p.svc == nil {
		panic(fmt.Errorf("in cephInterface.Get, p.svc = %v", p.svc))
	}

	initial := time.Now() //                        ***** Time starts
	head, err := p.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	latency := time.Since(initial) // 	        	***** Latency ends
	if err != nil {
		p.Printf("HeadObject err %v", err)  // FIXME log this
	}


	// https://gist.github.com/jboelter/ecfb08d6a18440ac16d93b5183aad207
	buff := &aws.WriteAtBuffer{}
	downloader := s3manager.NewDownloaderWithClient(p.svc)
	numBytes, err := downloader.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, head, fmt.Errorf("failed in downloader.Download, %v", err)
	}
	responseTime := time.Since(initial) - latency //  ***** Time ends
	if err != nil {
		rc := errorCodeToHTTPCode(err)
		fmt.Printf("%s %f %f 0 %d %s %d GET\n",
			initial.Format("2006-01-02 15:04:05.000"),
			latency.Seconds(), responseTime.Seconds(), numBytes, key, rc)
		// special case: non-success code from server  FIXME
		return buff.Bytes(), head, nil
	}
	fmt.Printf("%s %f %f 0 %d %s 200 GET\n",
		initial.Format("2006-01-02 15:04:05.000"),
		latency.Seconds(), responseTime.Seconds(), numBytes, key)
	return buff.Bytes(), head, nil
}

// Put puts a file and times it
// error return is used only by mkLoadTestFiles  FIXME
func (p S3Proto) Put(contents, path, bucket string) error {
	defer p.Begin("<contents>", path, bucket)()
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
