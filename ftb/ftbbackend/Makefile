default: docker push

linux:
	mkdir -p bin/
	GOOS=linux GOARCH=amd64 go build -o bin/ftbbackend_linux .

docker: linux
	docker build -t us.gcr.io/k8s-belgrade-workshop/ftbbackend:1.1 .

push:
	gcloud docker -- push us.gcr.io/k8s-belgrade-workshop/ftbbackend:1.1

darwin:
	go build -o bin/ftbbackend_darwin .

clean:
	rm -f bin/ftbbackend_linux bin/ftbbackend_darwin
