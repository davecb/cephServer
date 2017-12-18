install: prereqs
	go install github.com/davecb/cephServer/pkg/cephInterface
	go install github.com/davecb/cephServer/pkg/imageServer
	go install github.com/davecb/cephServer/pkg/trace
	go install github.com/davecb/cephServer/pkg/bucketServer
	go install github.com/davecb/cephServer/cmd/imager


prereqs:  ${HOME}/go/pkg/linux_amd64/github.com/nfnt/resize.a
	go get github.com/nfnt/resize
