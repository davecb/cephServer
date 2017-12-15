package bucketServer


import (
	"net/http"
	"fmt"
	"github.com/davecb/cephServer/pkg/trace"

	"strings"
	"github.com/davecb/cephServer/pkg/cephInterface"
	"github.com/aws/aws-sdk-go/service/s3"
)

var ceph *cephInterface.S3Proto   // maybe move

// Bucket is a rstorage bucket
type Bucket struct {
	trace.Trace
}

// New creates a bucket server
func New(t trace.Trace) *Bucket {
	ceph = cephInterface.New(t)
	return &Bucket{ t }
}

// Get tries to get an object from a specific bucket. Errors are wrotten to w
func (b Bucket) Get(w http.ResponseWriter, r *http.Request)  {  // nolint
	var head *s3.HeadObjectOutput
	defer b.Begin(r.URL.Path)()

	fmt.Fprintf(w, "got a request for %s\n", r.URL.Path)
	b.Printf("got a request for %s\n", r.URL.Path)

	// split path into bucket and url
	// FIXME handle bucket part of prefixes
	path := strings.TrimPrefix(r.URL.Path, "/content/v1/download.s3.kobo.com/")
	bytes, head, err := ceph.Get(path, "download.s3.kobo.com")
	if err != nil {
		http.Error(w, err.Error(), 999) // FIXME
	}  else {
		b.Printf("bucket.get worked, head = %v\n", head)
		w.Write(bytes) // nolint
	}

}


