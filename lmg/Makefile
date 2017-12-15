default: clean docker push

linux:
	mkdir -p bin/
	GOOS=linux GOARCH=amd64 go build -o bin/lmg_linux .

docker: linux
	docker build -t us.gcr.io/k8s-belgrade-workshop/lmg:1.1 .

push:
	gcloud docker -- push us.gcr.io/k8s-belgrade-workshop/lmg:1.1

darwin:
	go build -o bin/lmg_darwin .

clean:
	rm -f bin/hellok8s_linux bin/lmg_darwin
