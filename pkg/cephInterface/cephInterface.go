package cephInterface

import (
	"imageServer/pkg/trace"
	"os"
	"log"
	"strings"
)

// T is a debugging tool shared by the server components
var T trace.Trace

// P is the s3 protocol parameter set
var P = S3Proto{
	Prefix:   "nowhere",
	S3Bucket: "moose",
	Verbose:  true,
	S3Key:    "key",
	S3Secret: "secret",
}


// Get obtains a file from ceph via the s3-compatible interface
func Get(key string) (string, error) {
	defer T.Begin(key)()
	return "", nil
}

// Save stores a file in ceph via the s3-compatible interface
func Save(contents, key string) {
	defer T.Begin("<contents>", key)()

	strings := strings.Split(key, "/")
	out, err := os.Create(strings[len(strings)-1])
	if err != nil {
		log.Fatal(err)
	}
	_, err = out.WriteString(contents)
	if err != nil {
		log.Fatalf("write failure %v\n", err)
	}
	err = out.Close()
	if err != nil {
		log.Fatalf("write failure %v\n", err)
	}

}
