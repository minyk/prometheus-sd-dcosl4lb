PWD := `pwd`

default: build

build: clean
	docker run --rm -e "GO111MODULE=on" -e "GOOS=linux" -e "GOARCH=amd64" -e "CGO_ENABLED=0" -v $(PWD):/usr/src/github.com/minyk/prometheus-sd-dcosl4lb -w /usr/src/github.com/minyk/prometheus-sd-dcosl4lb golang:1.12 go build -a -tags netgo -ldflags="-s -w ${GO_LDFLAGS}" -v -o build/prometheus-sd-dcosl4lb-linux

clean:
	rm -rf ./build

docker: build
	docker build . -f docker/Dockerfile -t minyk/prometheus-dcosl4lb:v2.8.0
