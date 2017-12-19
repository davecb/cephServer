package bucketServer


import (
	"net/http"
	"github.com/davecb/cephServer/pkg/trace"
	"github.com/davecb/cephServer/pkg/cephInterface"
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
	var head map[string]string
	defer b.Begin(r.URL.Path)()

	b.Printf("got a request for %s\n", r.URL.Path)
	data, head, rc, err := ceph.Get(r.URL.Path, bucket)
	if err != nil {
		http.Error(w, err.Error(), 999) // FIXME
	}
	if rc != 200 {
		b.Printf("bucket.get failed, head = %v, rc = %d\n",
			head, rc )
		http.Error(w, err.Error(), rc)
	}
	b.Printf("bucket.get worked, head = %v\n", head)
	for key, value := range head {
		if value != "" {
			w.Header().Set(key, value)
		}
	}
	_, err = w.Write(data)
	if err != nil {
		b.Printf("oopsie! %v\n", err) // FIXME log this
	}

}


