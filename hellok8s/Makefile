default: docker push

linux:
	mkdir -p bin/
	GOOS=linux GOARCH=amd64 go build -o bin/hellok8s_linux .

docker: linux
	docker build -t us.gcr.io/k8s-belgrade-workshop/hellok8s:1.0 .

push:
	gcloud docker -- push us.gcr.io/k8s-belgrade-workshop/hellok8s:1.0

darwin:
	go build -o bin/hellok8s_darwin .

clean:
	rm -f bin/hellok8s_linux bin/hellok8s_darwin
