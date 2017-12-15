package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", helloHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalln("ERROR starting web server,", err)
	}
}

func helloHandler(w http.ResponseWriter, req *http.Request) {
	tmpl := template.New("page")
	var err error

	if tmpl, err = tmpl.Parse(tplHtml); err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	hostName, _ := os.Hostname()
	tmpl.Execute(w, hostName)
}

const tplHtml = `
<html>
<head>
	<title>Hello K8S</title>
</head>
<body>
	<br /><br /></br />
	<center><h1>Hello K8S</h1></center>
	<center><h3>Page served by: {{.}}</h3></center>
</body>
</html>
`
