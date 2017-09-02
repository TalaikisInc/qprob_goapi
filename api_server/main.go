package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/xenu256/qprob_goapi/api_server/database"
	"github.com/xenu256/qprob_goapi/api_server/models"
	"github.com/xenu256/qprob_goapi/api_server/v2handlers"

	"github.com/die-net/lrucache"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var cache = lrucache.New(104857600, 60*60*24) //100 Mb, 24 hours

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading environment variables.")
	}
}

func main() {
	ApiHost := os.Getenv("API_HOST")

	app := mux.NewRouter()
	app.Host(ApiHost)

	//API endpoints v1
	app.HandleFunc("/", RedirectHandler).Methods("GET")
	app.HandleFunc("/posts/", PostsHandler).Methods("GET")
	app.HandleFunc("/posts/{catSlug}/", PostsByCatHandler).Methods("GET")
	app.HandleFunc("/post/{postSlug}/", PostHandler).Methods("GET")
	app.HandleFunc("/today/", TodayHandler).Methods("GET")
	app.HandleFunc("/cats/", CategoriesHandler).Methods("GET")
	app.HandleFunc("/error/", ErrorHandler).Methods("GET")
	app.HandleFunc("/feed/", RssHandler).Methods("GET")

	//API endpoints v2
	app.HandleFunc("/v2.0/posts/{page}/", v2handlers.PostsHandler).Methods("GET")
	app.HandleFunc("/v2.0/cat/{catSlug}/{page}/", v2handlers.PostsByCatHandler).Methods("GET")
	app.HandleFunc("/v2.0/tag/{tag}/{page}/", v2handlers.PostsByTagHandler).Methods("GET")
	app.HandleFunc("/v2.0/today/{page}/", v2handlers.TodayPostsHandler).Methods("GET")
	app.HandleFunc("/v2.0/cal/{year}/{month}/{day}/{page}/", v2handlers.PostsByCalendarHandler).Methods("GET")
	app.HandleFunc("/v2.0/tags/{page}/", v2handlers.TagsHandler).Methods("GET")
	app.HandleFunc("/v2.0/cats/{page}/", v2handlers.CategoriesHandler).Methods("GET")
	app.HandleFunc("/v2.0/all_cats/", v2handlers.AllCategoriesHandler).Methods("GET")
	app.HandleFunc("/v2.0/top_cats/", v2handlers.TopCategoriesHandler).Methods("GET")
	app.HandleFunc("/v2.0/post/{postSlug}/", v2handlers.PostHandler).Methods("GET")
	app.HandleFunc("/v2.0/post_tags/{postSlug}/", v2handlers.TagsByPostHandler).Methods("GET")
	app.HandleFunc("/v2.0/popular/{hits}/{page}/", v2handlers.PopularPostsHandler).Methods("GET")
	app.HandleFunc("/v2.0/popular_posts/{page}/", v2handlers.PostsByPopularityHandler).Methods("GET")
	app.HandleFunc("/v2.0/most_popular/{page}/", v2handlers.MostPopularPostsHandler).Methods("GET")
	app.HandleFunc("/v2.0/filled_tags/{cnt}/", v2handlers.FilledTagsHandler).Methods("GET")
	app.HandleFunc("/v2.0/top_tags/", v2handlers.TopTagsHandler).Methods("GET")
	app.HandleFunc("/v2.0/meta/", v2handlers.MetaHandler).Methods("GET")
	app.HandleFunc("/v2.0/sentiment/", v2handlers.SentimentHandler).Methods("GET")

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
		Addr:         ApiHost + ":" + os.Getenv("API_PORT"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(server.ListenAndServe())

}

func RssHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")

	Host := os.Getenv("HOST")
	domain := "https://" + Host + "/"
	Author := os.Getenv("SHORT_SITE_NAME")
	SiteName := os.Getenv("SITE_NAME")

	cached, isCached := cache.Get("rss")
	if isCached == false {
		type Item struct {
			Title   string `xml:"title"`
			Link    string `xml:"link"`
			Author  string `xml:"author"`
			Created string `xml:"date"`
		}

		db := database.Connect()
		defer db.Close()

		dateBack := time.Now().AddDate(0, 0, -2)

		query := fmt.Sprintf(`SELECT 
			posts.title, 
			CONCAT('%[1]s', posts.slug, '/'), 
			cats.title, 
			posts.date 
			FROM aggregator_post as posts 
			INNER JOIN aggregator_category as cats ON posts.category_id = cats.title 
			WHERE date > '%[2]s' 
			ORDER BY date DESC;`, domain, dateBack)
		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		posts := make([]Item, 0)
		for rows.Next() {
			post := Item{}
			err := rows.Scan(&post.Title, &post.Link, &post.Author, &post.Created)
			if err != nil {
				return
			}
			posts = append(posts, post)
		}
		if err = rows.Err(); err != nil {
			return
		}

		rss, err := xml.MarshalIndent(posts, "", "    ")
		if err != nil {
			return
		}

		Host := os.Getenv("HOST")
		Author := os.Getenv("SHORT_SITE_NAME")
		SiteName := os.Getenv("SITE_NAME")

		uEl := []byte("<channel>" +
			"<title>" + SiteName + "</title>" +
			"<author>" + Author + "</author>" +
			"<link>https://" + Host + "</link>")
		dEl := []byte("</channel>")

		// FIXME by some reason non-cached version doesn't produce correct xml
		w.Write(uEl)
		w.Write([]byte(rss))
		w.Write(dEl)

		cache.Set("rss", rss)

	}

	uEl := []byte("<channel>" +
		"<title>" + SiteName + "</title>" +
		"<author>" + Author + "</author>" +
		"<link>https://" + Host + "</link>")
	dEl := []byte("</channel>")

	w.Write(uEl)
	w.Write(cached)
	w.Write(dEl)
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

func ErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
}
