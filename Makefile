test: install
	cd objectServer; go test


install: prereqs
	go install github.com/davecb/cephServer/pkg/cephInterface
	#go install github.com/davecb/cephServer/pkg/imageServer
	go install github.com/davecb/trace
	go install github.com/davecb/cephServer/pkg/objectServer
	#go install github.com/davecb/cephServer/cmd/imager


#prereqs:  ${HOME}/go/pkg/linux_amd64/github.com/nfnt/resize.a
#	go get github.com/nfnt/resize
#
prereqs:  
	go get github.com/aws/aws-sdk-go/aws
	go get github.com/davecb/trace
