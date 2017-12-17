package bucketServer


import (
	"net/http"
	"github.com/davecb/cephServer/pkg/trace"
	"github.com/davecb/cephServer/pkg/cephInterface"
	"github.com/aws/aws-sdk-go/service/s3"
)

var ceph *cephInterface.S3Proto   // maybe move

// Bucket is a storage bucket
type Bucket struct {
	trace.Trace
}

// New creates a bucket server
func New(t trace.Trace) *Bucket {
	ceph = cephInterface.New(t)
	return &Bucket{ t }
}

// Get an object from a specific bucket. Errors are written to w
func (b Bucket) Get(w http.ResponseWriter, r *http.Request, bucket string)  {  // nolint
	var head *s3.HeadObjectOutput
	defer b.Begin(r.URL.Path)()

	b.Printf("got a request for %s\n", r.URL.Path)
	bytes, head, err := ceph.Get(r.URL.Path, bucket)
	if err != nil {
		http.Error(w, err.Error(), 999) // FIXME
	}  else {
		b.Printf("bucket.get worked, head = %v\n", head)
		w.Write(bytes) // nolint
	}

}


