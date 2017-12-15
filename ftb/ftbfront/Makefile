default: clean docker push

linux:
	mkdir -p bin/
	GOOS=linux GOARCH=amd64 go build -o bin/ftbfront_linux .

docker: linux
	docker build -t us.gcr.io/k8s-belgrade-workshop/ftbfront:1.2 .

push:
	gcloud docker -- push us.gcr.io/k8s-belgrade-workshop/ftbfront:1.2

darwin:
	go build -o bin/ftbfront_darwin .

clean:
	rm -f bin/ftbfront_linux bin/ftbfront_darwin
