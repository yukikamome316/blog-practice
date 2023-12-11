package main

import (
	"log"
	"net/http"
	"text/template"
)

const templatePath = "./templates"
const layoutPath = templatePath + "/layout.html"

var (
	templates = template.Must(template.ParseFiles(layoutPath, templatePath+"/index.html"))
)

func main() {
	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "layout.html", map[string]interface{}{
		"PageTitle": "Home Page",
		"Text":      "Hello World",
	})
}
