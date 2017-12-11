install: prereqs
    go install cephServer/pkg/cephInterface
    go install cephServer/pkg/imageMigrator
    go install cephServer/pkg/imageResizer
    go install cephServer/pkg/trace
    go install cephServer/cmd/imager


prereqs:  ${HOME}/go/pkg/linux_amd64/github.com/nfnt/resize.a
    go get github.com/nfnt/resize
