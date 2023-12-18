package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
	templatePath = "./templates"
	layoutPath   = templatePath + "/layout.html"
	createPath   = templatePath + "/create.html"

	dbPath = "./db.sqlite3"

	createPostTableQuery = `CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			body TEXT,
			author TEXT,
			created_at INTEGER
		)`

	// ブログポストテーブルにデータを挿入するSQL文
	insertPostQuery = `INSERT INTO posts (title, body, author, created_at) VALUES (?, ?, ?, ?)`
)

var (
	db             *sqlx.DB
	indexTemplate  = template.Must(template.ParseFiles(layoutPath, templatePath+"/index.html"))
	createTemplate = template.Must(template.ParseFiles(layoutPath, createPath))
)

func main() {
	db = dbConnect()
	defer db.Close()
	err := initDB()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/post/new", createPostHandler)

	fmt.Println("Server is listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	indexTemplate.ExecuteTemplate(w, "layout.html", map[string]interface{}{
		"PageTitle": "Home Page",
		"Text":      "Hello World",
	})
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// GETリクエストの場合はテンプレートを表示
		createTemplate.ExecuteTemplate(w, "layout.html", map[string]interface{}{
			"PageTitle": "ブログポスト作成",
		})
	} else if r.Method == "POST" {
		// POSTリクエストの場合はブログポストを作成
		title := r.FormValue("title")
		body := r.FormValue("body")
		author := r.FormValue("author")
		createdAt := time.Now().Unix()
		// フォームに空の項目がある場合はエラーを返す
		if title == "" || body == "" || author == "" {
			log.Print("フォームに空の項目があります")
			createTemplate.ExecuteTemplate(w, "layout.html", map[string]interface{}{
				"Message": "フォームに空の項目があります",
			})
			return
		}
		err := insertPost(title, body, author, createdAt)
		if err != nil {
			log.Print(err)
			return
		}
		// 作成したブログポストを表示
		http.Redirect(w, r, "/post", http.StatusMovedPermanently)
	}
}

func insertPost(title string, body string, author string, createdAt int64) error {
	// ブログポストテーブルにデータを挿入　last_insert_rowid()で最後に挿入したデータのIDを取得
	_, err := db.Exec(insertPostQuery, title, body, author, createdAt)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return err
	}
	return nil
}

func dbConnect() *sqlx.DB {
	// SQLite3のデータベースに接続
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func initDB() error {
	_, err := db.Exec(createPostTableQuery)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}
