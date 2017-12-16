default: clean docker push

linux:
	mkdir -p bin/
	GOOS=linux GOARCH=amd64 go build -o bin/demo1_linux .

docker: linux
	docker build -t us.gcr.io/k8s-belgrade-workshop/demo1:1.7 .

push:
	gcloud docker -- push us.gcr.io/k8s-belgrade-workshop/demo1:1.7

darwin:
	go build -o bin/demo1_darwin .

clean:
	rm -f bin/demo1_linux bin/demo1_darwin

