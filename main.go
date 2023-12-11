package main

import (
	"log"
	"net/http"
	"text/template"
)

const templatePath = "./templates"
const layoutPath = templatePath + "/layout.html"

var (
	templates = template.Must(template.ParseGlob(templatePath + "/index.html"))
)

func main() {
	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Title": "Hello World",
		"Text":  "Hello World",
	})
}
