package bucketServers


import (
	//ceph "cephServer/pkg/cephInterface"

	"net/http"
	"fmt"
)

// Get gets an object from a specific bucket
func Get(w http.ResponseWriter, r *http.Request) error {
	defer T.Begin(r.URL.Path)()

	fmt.Fprintf(w, "got a request for %s\n", r.URL.Path)
	return nil

	//// split path into bucket and url
	//// FIXME handle bucket part of prefixes
	//path := strings.TrimPrefix(r.URL.Path, "/content/v1/download.s3.kobo.com/")
	//bytes, err := ceph.Get(path, "download.s3.kobo.com")
	//if err == nil {
	//	return err
	//}
	//io.WriteString(w, bytes) // nolint
	//return nil
}


