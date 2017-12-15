package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var backendHost = os.Getenv("CUSTOM_BACKEND_HOST")

func main() {
	log.SetOutput(os.Stdout)
	log.Println("Starting front app with backend host:", backendHost)
	http.HandleFunc("/", rootHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Println("ERROR starting web server,", err)
	}
}

const pageHtml = `
<html>
<head>
	<title>Front-To-Back</title>
	<link rel="stylesheet" type="text/css" href="/static/style.css">
</head>
<body>
	<br /><br /></br />
	<center><h1>Front-To-Back.. in style</h1></center><hr />
	<center><p>Magic number is: {{.Number}}</p></center>
	<center><p>Calculated by: {{.Host}}</p></center>
</body>
</html>
`

func rootHandler(w http.ResponseWriter, req *http.Request) {
	tmpl := template.New("page")
	var err error
	if tmpl, err = tmpl.Parse(pageHtml); err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	resp, err := http.Get("http://" + backendHost + "/api/random")
	if err != nil {
		tmpl.Execute(w, "unknown - "+err.Error())
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		tmpl.Execute(w, "unknown - "+err.Error())
		return
	}
	log.Println("raw data response:", string(data))

	var randomResponse randomResponseType
	json.Unmarshal(data, &randomResponse)
	log.Printf("parsed response data: %+v\n", randomResponse)
	tmpl.Execute(w, randomResponse)

}

type randomResponseType struct {
	Number int64  `json:"number"`
	Host   string `json:"host"`
}
