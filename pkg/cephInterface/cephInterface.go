package cephInterface

import (
	"github.com/davecb/cephServer/pkg/trace"
)

// T is a debugging tool shared by the server components
var T trace.Trace

// P is the s3 protocol parameter set
var P = S3Proto{
	Prefix:   "nowhere",
	Verbose:  true,
	S3Key:    "key",
	S3Secret: "secret",
}


// Get obtains a file from ceph via the s3-compatible interface
func Get(key, bucket string) (string, error) {
	defer T.Begin(key, bucket)()
	//contents, err := P.Get(key, bucket)
	//if err != nil {
	//	return "", fmt.Errorf("died getting %s %s, %v\n",key, bucket, err)
	//}
	//return contents, nil
	return "fake data", nil
}

// Save stores a file in ceph via the s3-compatible interface
func Put(contents, key, bucket string) error {
	defer T.Begin("<contents>", key, bucket)()

	//err := P.Put(contents, key, bucket)
	//if err != nil {
	//	return fmt.Errorf("died putting %s %s, %v\n",key, bucket, err)
	//}
	return nil
}
