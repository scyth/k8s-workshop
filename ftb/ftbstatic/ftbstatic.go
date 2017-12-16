package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/static/style.css", styleHandler)
	http.HandleFunc("/", rootHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Println("ERROR starting web server,", err)
	}
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%s", "go away")
}

func styleHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%s", tplHtml)
}

const tplHtml = `
h1 {
    color: red;
}
`
