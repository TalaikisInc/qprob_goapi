package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"./database"
	"./models"
	"./v2handlers"

	"github.com/die-net/lrucache"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var cache = lrucache.New(104857600, 10800) //100 Mb, 3 hours

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading environment variables.")
	}
}

func main() {
	Host := os.Getenv("API_HOST")

	app := mux.NewRouter()
	app.Host(Host)

	//API endpoints v1
	app.HandleFunc("/", RedirectHandler)
	app.HandleFunc("/posts/", PostsHandler)
	app.HandleFunc("/posts/{catSlug}/", PostsByCatHandler)
	app.HandleFunc("/post/{postSlug}/", PostHandler)
	app.HandleFunc("/today/", TodayHandler)
	app.HandleFunc("/cats/", CategoriesHandler)

	//API endpoints v2
	app.HandleFunc("/v2.0/posts/{page}/", v2handlers.PostsHandler)
	app.HandleFunc("/v2.0/cat/{catSlug}/{page}/", v2handlers.PostsByCatHandler)
	app.HandleFunc("/v2.0/tag/{tag}/{page}/", v2handlers.PostsByTagHandler)
	app.HandleFunc("/v2.0/today/{page}/", v2handlers.TodayPostsHandler)
	app.HandleFunc("/v2.0/cal/{year}/{month}/{day}/{page}/", v2handlers.PostsByCalendarHandler)
	app.HandleFunc("/v2.0/tags/{page}/", v2handlers.TagsHandler)
	app.HandleFunc("/v2.0/cats/{page}/", v2handlers.CategoriesHandler)
	app.HandleFunc("/v2.0/post/{postSlug}/", v2handlers.PostHandler)
	app.HandleFunc("/v2.0/post_tags/{postSlug}/", v2handlers.TagsByPostHandler)
	app.HandleFunc("/v2.0/popular/{page}/", v2handlers.PopularHandler)
	app.HandleFunc("/v2.0/filled_tags/{cnt}/", v2handlers.FilledTagsHandler)

	//TODOs:
	//app.HandleFunc("/tweets/{postId}/", TweetsByPostHandler)
	//app.HandleFunc("/videos/{postId}/", VideosByPostHandler)
	//app.HandleFunc("/tweets/{tag}/", TweetsByTagHandler)
	//app.HandleFunc("/videos/{tag}/", VideosByTagtHandler)
	//app.HandleFunc("/books/{postId}/", BooksByPostHandler)
	//app.HandleFunc("/books/{tag}/", BooksByTagHandler)
	//app.HandleFunc("/posts/{query}", SearchHandler)
	//add source form
	//feedback form

	server := &http.Server{
		Handler:      app,
		Addr:         Host + ":" + os.Getenv("API_PORT"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(server.ListenAndServe())

}

func PostsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cached, isCached := cache.Get("posts")
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := `SELECT posts.title, posts.slug, posts.url, posts.summary, posts.date, 
		posts.sentiment, COALESCE(posts.image, ""), posts.category_id, cats.slug, 
		COALESCE(cats.thumbnail, "") FROM aggregator_post as posts 
		INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title 
		ORDER BY date DESC  LIMIT 100;`
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary,
				&post.Date, &post.Sentiment, &post.Image, &post.CategoryID.Title,
				&post.CategoryID.Slug, &post.CategoryID.Thumbnail)
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
			log.Fatal(err)
		}

		cache.Set("posts", j)
		w.Write(j)
	}
	w.Write(cached)
}

func CategoriesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cached, isCached := cache.Get("cats")
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := `SELECT cats.title, cats.slug, COALESCE(cats.thumbnail, ""), 
		count(posts.title) AS cnt FROM aggregator_category AS cats INNER JOIN 
		aggregator_post AS posts ON posts.category_id = cats.title 
		GROUP BY cats.title;`

		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		categories := make([]models.Category, 0)
		for rows.Next() {
			category := models.Category{}

			err := rows.Scan(&category.Title, &category.Slug, &category.Thumbnail,
				&category.PostCnt)
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
			log.Fatal(err)
		}

		cache.Set("cats", j)
		w.Write(j)
	}
	w.Write(cached)
}

func PostsByCatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cat := url.QueryEscape(strings.Split(r.RequestURI, "/")[2])

	cached, isCached := cache.Get("posts_cat_" + cat)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT posts.title, posts.slug, posts.url, posts.summary, 
		posts.date AS dt, posts.sentiment, COALESCE(posts.image, ""), posts.category_id, 
		cats.slug, COALESCE(cats.thumbnail, "") FROM aggregator_post 
		AS posts INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title WHERE 
		cats.slug='%s' ORDER BY dt DESC LIMIT 100;`, cat)
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary,
				&post.Date, &post.Sentiment, &post.Image, &post.CategoryID.Title,
				&post.CategoryID.Slug, &post.CategoryID.Thumbnail)
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
			log.Fatal(err)
		}

		cache.Set("posts_cat_"+cat, j)
		w.Write(j)
	}
	w.Write(cached)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	postSlug := url.QueryEscape(strings.Split(r.RequestURI, "/")[2])

	cached, isCached := cache.Get("post_" + postSlug)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT posts.title, posts.slug, posts.url, posts.summary, 
		posts.date, posts.sentiment, COALESCE(posts.image, ""), posts.category_id, 
		cats.slug, COALESCE(cats.thumbnail, "") FROM aggregator_post as posts 
		INNER JOIN aggregator_category as cats ON posts.category_id = cats.title WHERE 
		posts.slug='%s';`, postSlug)
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary,
				&post.Date, &post.Sentiment, &post.Image, &post.CategoryID.Title,
				&post.CategoryID.Slug, &post.CategoryID.Thumbnail)
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
			log.Fatal(err)
		}

		cache.Set("post_"+postSlug, j)
		w.Write(j)
	}
	w.Write(cached)
}

func TodayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cached, isCached := cache.Get("today")
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		dateBack := time.Now().AddDate(0, 0, -2) //2 days back for "today"

		query := fmt.Sprintf(`SELECT posts.title, posts.slug, posts.url, posts.summary, 
		posts.date, posts.sentiment, COALESCE(posts.image, ""), posts.category_id, 
		cats.slug, COALESCE(cats.thumbnail, "") FROM aggregator_post as posts INNER JOIN 
		aggregator_category as cats ON posts.category_id = cats.title 
		WHERE date > '%s' ORDER BY date DESC;`, dateBack)

		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary,
				&post.Date, &post.Sentiment, &post.Image, &post.CategoryID.Title,
				&post.CategoryID.Slug, &post.CategoryID.Thumbnail)
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
			log.Fatal(err)
		}

		cache.Set("today", j)
		w.Write(j)
	}
	w.Write(cached)
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
		log.Fatal(err)
	}
}
