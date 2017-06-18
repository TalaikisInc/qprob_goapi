package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"./database"
	"./models"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func init() {
	//note, this package isn't fully compatible with pydotenv!
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading environment variables.")
	}
}

func main() {
	app := mux.NewRouter()

	app.HandleFunc("/", RedirectHandler)
	app.HandleFunc("/posts/", PostsHandler)
	app.HandleFunc("/posts/{catSlug}/", PostsByCatHandler)
	app.HandleFunc("/post/{postSlug}/", PostHandler)
	app.HandleFunc("/today/", TodayHandler)
	app.HandleFunc("/cats/", CategoryHandler)
	//TODOs:
	//app.HandleFunc("/tweets/", HomeHandler)
	//app.HandleFunc("/videos/", HomeHandler)
	//app.HandleFunc("/books/", HomeHandler)
	//app.HandleFunc("/posts/{query}", SearchHandler)

	server := &http.Server{
		Handler:      app,
		Addr:         os.Getenv("DEV_API_HOST"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

func PostsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db := database.Connect()
	defer db.Close()

	query := `SELECT posts.title, posts.slug, posts.url, posts.summary, posts.date, posts.sentiment, posts.image, 
		posts.category_id, cats.slug FROM aggregator_post as posts INNER JOIN aggregator_category AS cats 
		ON posts.category_id = cats.title ORDER BY date DESC  LIMIT 100;`
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 404)
	}
	defer rows.Close()

	posts := make([]models.Post, 0)
	for rows.Next() {
		post := models.Post{}
		err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date,
			&post.Sentiment, &post.Image, &post.CategoryID.Title, &post.CategoryID.Slug)
		if err != nil {
			panic(err)
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}

	j, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
	}
	w.Write(j)
}

func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db := database.Connect()
	defer db.Close()

	query := `SELECT title, slug FROM aggregator_category;`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	categories := make([]models.Category, 0)
	for rows.Next() {
		category := models.Category{}
		err := rows.Scan(&category.Title, &category.Slug)
		if err != nil {
			panic(err)
		}
		categories = append(categories, category)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}

	j, err := json.Marshal(categories)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
	}
	w.Write(j)
}

func PostsByCatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := database.Connect()
	defer db.Close()

	query := fmt.Sprintf(`SELECT posts.title, posts.slug, posts.url, posts.summary, posts.date AS 
		dt, posts.sentiment, posts.image, posts.category_id, cats.slug AS cat FROM aggregator_post 
		AS posts INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title WHERE 
		cats.slug='%s' ORDER BY dt DESC LIMIT 100;`, strings.Split(r.RequestURI, "/")[2])
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 404)
	}
	defer rows.Close()

	posts := make([]models.Post, 0)
	for rows.Next() {
		post := models.Post{}
		err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date,
			&post.Sentiment, &post.Image, &post.CategoryID.Title, &post.CategoryID.Slug)
		if err != nil {
			panic(err)
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}

	j, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
	}
	w.Write(j)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db := database.Connect()
	defer db.Close()

	query := fmt.Sprintf(`SELECT posts.title, posts.slug, posts.url, posts.summary, posts.date, 
		posts.sentiment, posts.image, posts.category_id, cats.slug FROM aggregator_post as posts 
		INNER JOIN aggregator_category as cats ON posts.category_id = cats.title WHERE 
		posts.slug='%s';`, strings.Split(r.RequestURI, "/")[2])
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 404)
	}
	defer rows.Close()

	posts := make([]models.Post, 0)
	for rows.Next() {
		post := models.Post{}
		err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date, &post.Sentiment,
			&post.Image, &post.CategoryID.Title, &post.CategoryID.Slug)
		if err != nil {
			panic(err)
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}

	j, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
	}
	w.Write(j)
}

func TodayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db := database.Connect()
	defer db.Close()

	dateBack := time.Now().AddDate(0, 0, -2) //2 days back for "today"
	fmt.Println(dateBack)

	query := fmt.Sprintf(`SELECT posts.title, posts.slug, posts.url, posts.summary, posts.date, posts.sentiment, 
		posts.image, posts.category_id, cats.slug FROM aggregator_post as posts INNER JOIN 
		aggregator_category as cats ON posts.category_id = cats.title WHERE date > '%s' ORDER BY date DESC;`, dateBack)

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 404)
	}
	defer rows.Close()

	posts := make([]models.Post, 0)
	for rows.Next() {
		post := models.Post{}
		err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date, &post.Sentiment, &post.Image,
			&post.CategoryID.Title, &post.CategoryID.Slug)
		if err != nil {
			panic(err)
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}

	j, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
	}
	w.Write(j)
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	type Page struct {
		RedirectTitle string
		RedirectUrl   string
	}

	tpl := template.Must(template.ParseFiles("templates/redirect.html"))

	err := tpl.Execute(w, Page{
		RedirectTitle: os.Getenv("API_REDIRECT_TITLE"),
		RedirectUrl:   string(os.Getenv("API_DESCRIPTION_URL"))})
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
	}
}
