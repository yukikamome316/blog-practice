package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
	templatePath = "./templates"
	layoutPath   = templatePath + "/layout.html"
	createPath   = templatePath + "/create.html"
	postPath     = templatePath + "/post.html"
	editPath     = templatePath + "/edit.html"

	publicPath = "./public"
	dbPath     = "./db.sqlite3"

	createPostTableQuery = `CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			body TEXT,
			author TEXT,
			created_at INTEGER
		)`

	insertPostQuery     = `INSERT INTO posts (title, body, author, created_at) VALUES (?, ?, ?, ?)`
	selectPostByIdQuery = `SELECT * FROM posts WHERE id = ?`
	selectAllPostsQuery = `SELECT * FROM posts`
	deletePostQuery     = `DELETE FROM posts WHERE id = ?`
	updatePostQuery     = `UPDATE posts SET title = ?, body = ?, author = ?, created_at = ? WHERE id = ?`
)

var (
	db *sqlx.DB

	funcDate = template.FuncMap{
		"date": func(t int64) string {
			return time.Unix(t, 0).Format("2006-01-02 15:04:05")
		},
	}

	indexTemplate  = template.Must(template.New("layout.html").Funcs(funcDate).ParseFiles(layoutPath, templatePath+"/index.html"))
	createTemplate = template.Must(template.ParseFiles(layoutPath, createPath))
	postTemplate   = template.Must(template.ParseFiles(layoutPath, postPath))
	editTemplate   = template.Must(template.ParseFiles(layoutPath, editPath))
)

type Post struct {
	ID        int    `db:"id"`
	Title     string `db:"title"`
	Body      string `db:"body"`
	Author    string `db:"author"`
	CreatedAt int64  `db:"created_at"`
}

func main() {
	db = dbConnect()
	defer db.Close()
	err := initDB()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/post/", postHandler)
	http.HandleFunc("/post/new", createPostHandler)
	http.HandleFunc("/post/delete/", deletePostHandler)
	http.HandleFunc("/post/edit/", editPostHandler)

	// cssファイルを配信
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(publicPath+"/css"))))

	fmt.Println("Server is listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := getAllPosts()
	if err != nil {
		log.Print(err)
		return
	}

	indexTemplate.ExecuteTemplate(w, "layout.html", map[string]interface{}{
		"PageTitle": "記事一覧",
		"Posts":     posts,
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
		id, err := insertPost(title, body, author, createdAt)
		if err != nil {
			log.Print(err)
			return
		}
		// 作成したブログポストを表示
		http.Redirect(w, r, "/post/"+strconv.FormatInt(id, 10), http.StatusMovedPermanently)
	}
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {
	// URLのPathからIDを取得
	id := r.URL.Path[len("/post/delete/"):]
	//idをint型に変換
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return
	}
	// ブログポストを取得
	err = deletePostById(idInt)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return
	}
	// テンプレートを表示
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func editPostHandler(w http.ResponseWriter, r *http.Request) {
	// URLのPathからIDを取得
	id := r.URL.Path[len("/post/edit/"):]
	//idをint型に変換
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return
	}
	// ブログポストを取得
	post, err := getPostById(idInt)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return
	}
	if r.Method == "GET" {
		// GETリクエストの場合はテンプレートを表示
		editTemplate.ExecuteTemplate(w, "layout.html", map[string]interface{}{
			"PageTitle": "ブログポスト編集",
			"ID":        post.ID,
			"Title":     post.Title,
			"Body":      post.Body,
			"Author":    post.Author,
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
			editTemplate.ExecuteTemplate(w, "layout.html", map[string]interface{}{
				"Message": "フォームに空の項目があります",
			})
			return
		}
		err := updatePostById(idInt, title, body, author, createdAt)
		if err != nil {
			log.Print(err)
			return
		}
		// 作成したブログポストを表示
		http.Redirect(w, r, "/post/"+strconv.Itoa(idInt), http.StatusMovedPermanently)
	}
}

func dbConnect() *sqlx.DB {
	// SQLite3のデータベースに接続
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	// URLのPathからIDを取得
	id := r.URL.Path[len("/post/"):]
	//idをint型に変換
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return
	}
	// ブログポストを取得
	post, err := getPostById(idInt)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return
	}
	// テンプレートを表示
	postTemplate.ExecuteTemplate(w, "layout.html", map[string]interface{}{
		"Title":     post.Title,
		"ID":        post.ID,
		"PageTitle": post.Title,
		"Body":      post.Body,
		"CreatedAt": time.Unix(post.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		"Author":    post.Author,
	})
}

func updatePostById(id int, title string, body string, author string, createdAt int64) error {
	_, err := db.Exec(updatePostQuery, title, body, author, createdAt, id)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

// 全てのブログポストを取得
func getAllPosts() ([]Post, error) {
	// ブログポストを全て取得
	var posts []Post
	err := db.Select(&posts, selectAllPostsQuery)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return posts, err
	}
	return posts, nil
}

// ブログポストをIDで取得
func getPostById(id int) (Post, error) {
	// ブログポストを取得
	var post Post
	err := db.Get(&post, selectPostByIdQuery, id)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return post, err
	}
	return post, nil
}

// ブログポストを作成
func insertPost(title string, body string, author string, createdAt int64) (int64, error) {
	// ブログポストテーブルにデータを挿入　last_insert_rowid()で最後に挿入したデータのIDを取得
	result, err := db.Exec(insertPostQuery, title, body, author, createdAt)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return 0, err
	}
	println(result)
	id, err := result.LastInsertId()
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return 0, err
	}
	return id, nil
}

func initDB() error {
	_, err := db.Exec(createPostTableQuery)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func deletePostById(id int) error {
	_, err := db.Exec(deletePostQuery, id)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}
