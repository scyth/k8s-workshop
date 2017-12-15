default: docker push

linux:
	mkdir -p bin/
	GOOS=linux GOARCH=amd64 go build -o bin/ftbstatic_linux .

docker: linux
	docker build -t us.gcr.io/k8s-belgrade-workshop/ftbstatic:1.0 .

push:
	gcloud docker -- push us.gcr.io/k8s-belgrade-workshop/ftbstatic:1.0

darwin:
	go build -o bin/ftbstatic_darwin .

clean:
	rm -f bin/ftbstatic_linux bin/ftbstatic_darwin
