package main

import (
	"database/sql"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var (
	mysqlHost   = os.Getenv("MYSQL_SERVICE_HOST")
	mysqlPort   = os.Getenv("MYSQL_SERVICE_PORT")
	mysqlUser   = os.Getenv("MYSQL_CLIENT_USERNAME")
	mysqlPass   = os.Getenv("MYSQL_CLIENT_PASSWORD")
	mysqlDbname = os.Getenv("MYSQL_CLIENT_DBNAME")

	db *sql.DB
)

type Person struct {
	Id        int64
	FirstName string
	LastName  string
}

func main() {
	log.SetOutput(os.Stdout)
	log.Printf("LMG: mysqlHost:%s, mysqlDbname:%s, mysqlUser:%s, mysqlPass:%s\n", mysqlHost, mysqlDbname, mysqlUser, mysqlPass)

	var err error
	db, err = sql.Open("mysql", mysqlUser+":"+mysqlPass+"@("+mysqlHost+":"+mysqlPort+")/"+mysqlDbname)
	if err != nil {
		log.Fatalln("LMG: failed to start mysql connection,", err.Error())
	}
	defer db.Close()

	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":8080", nil)
}

type pageData struct {
	HaveError bool
	Error     error
	People    []*Person
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	tmpl := template.New("page")
	var err error
	if tmpl, err = tmpl.Parse(pageHtml); err != nil {
		http.Error(w, "Error parsing template, "+err.Error(), http.StatusInternalServerError)
		return
	}

	people, err := getPeople(db)
	if err != nil {
		tmpl.Execute(w, &pageData{true, err, nil})
		return
	}
	tmpl.Execute(w, &pageData{false, nil, people})
}

func getPeople(db *sql.DB) ([]*Person, error) {
	rows, err := db.Query("select id, first_name, last_name from people order by id")
	if err != nil {
		return nil, err
	}

	results := make([]*Person, 0)

	for rows.Next() {
		var record Person
		if err = rows.Scan(&record.Id, &record.FirstName, &record.LastName); err != nil {
			rows.Close()
			return nil, errors.New("row scan failed: " + err.Error())
		}
		results = append(results, &record)
	}
	rows.Close()
	return results, nil
}

const pageHtml = `
<html>
<head>
	<title>LMG</title>
</head>
<body>
	<br /><br /></br />
	<center><h1>LMG - People List</h1></center><hr />

	{{if .HaveError}}
		Failed to fetch data from MySQL: {{.Error}}
	{{else}}
		<center>
		{{if not .People}}
		No data found
		{{else}}
		<table border="1" width="500">
			<tr>
				<th width=>ID</th>
				<th>FirstName</th>
				<th>LastName</th>
			</tr>
			{{range .People}}
			<tr>
				<td>{{.Id}}</td>
				<td>{{.FirstName}}</td>
				<td>{{.LastName}}</td>
			</tr>
			{{end}}
		</table>
		{{end}}
		</center>
	{{end}}
</body>
</html>
`
