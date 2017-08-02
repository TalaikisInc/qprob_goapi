package v2handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"../database"
	"../models"
	"github.com/die-net/lrucache"
)

var cache = lrucache.New(104857600*3, 10800*2) //300 Mb, 6 hours
var postsPerPage = 20
var catsPerPage = 100

func TagsByPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	title := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])

	cached, isCached := cache.Get("post_tags_" + title)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT  
			tags.title, 
			tags.slug 
			FROM aggregator_tags AS tags 
			INNER JOIN aggregator_post_tags AS post_tags ON tags.title = post_tags.tags_id 
			INNER JOIN aggregator_post as posts ON post_tags.post_id = posts.title
			WHERE posts.slug='%[1]s';`, title)

		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		tags := make([]models.Tag, 0)
		for rows.Next() {
			tag := models.Tag{}

			err := rows.Scan(&tag.Title, &tag.Slug)
			if err != nil {
				panic(err)
			}
			tags = append(tags, tag)
		}
		if err = rows.Err(); err != nil {
			panic(err)
		}

		j, err := json.Marshal(tags)
		if err != nil {
			log.Fatal(err)
		}

		cache.Set("post_tags_"+title, j)
		w.Write(j)

	}
	w.Write(cached)
}

func PostsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	p, err := strconv.Atoi(page)
	if err != nil {
		log.Fatal(err)
	}

	cached, isCached := cache.Get("posts_" + page)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT 
			posts.title, 
			posts.slug, 
			posts.url, 
			posts.summary, 
			posts.date, 
			posts.sentiment, 
			COALESCE(posts.image, ""), 
			posts.category_id, 
			cats.slug, 
			COALESCE(cats.thumbnail, "") 
			FROM aggregator_post as posts
			INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title 
			ORDER BY posts.date DESC 
			LIMIT %[1]d,%[2]d;`, postsPerPage*p, postsPerPage)

		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date,
				&post.Sentiment, &post.Image, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail)
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

		cache.Set("posts_"+page, j)
		w.Write(j)
	}
	w.Write(cached)
}

func PostsHandler2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	p, err := strconv.Atoi(page)
	if err != nil {
		log.Fatal(err)
	}

	cached, isCached := cache.Get("posts2_" + page)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT
			posts.title,
			posts.slug, 
			posts.url, 
			posts.summary, 
			posts.date, 
			posts.sentiment, 
			COALESCE(posts.image, ""), 
			posts.category_id, 
			cats.slug, 
			COALESCE(cats.thumbnail, "") 
			FROM (
				SELECT title
					FROM aggregator_post 
					ORDER BY date DESC 
					LIMIT %[1]d,%[2]d 
				) AS limited 
			JOIN aggregator_post AS posts ON posts.title = limited.title 
			INNER JOIN aggregator_category AS cats ON limited.category_id = cats.title 
			ORDER BY posts.title;`, postsPerPage*p, postsPerPage)

		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date,
				&post.Sentiment, &post.Image, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail)
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

		cache.Set("posts2_"+page, j)
		w.Write(j)
	}
	w.Write(cached)
}

func PostsByCatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cat := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[4])
	p, err := strconv.Atoi(page)
	if err != nil {
		log.Fatal(err)
	}

	cached, isCached := cache.Get("posts_cat_" + cat + "_" + page)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT 
			posts.title, 
			posts.slug, 
			posts.url, 
			posts.summary, 
			posts.date AS dt, 
			posts.sentiment, 
			COALESCE(posts.image, ""), 
			posts.category_id, 
			cats.slug, 
			COALESCE(cats.thumbnail, "") 
			FROM aggregator_post AS posts 
			INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title 
			WHERE cats.slug='%[1]s' 
			ORDER BY dt DESC 
			LIMIT %[2]d,%[3]d;`, cat, postsPerPage*p, postsPerPage)
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date,
				&post.Sentiment, &post.Image, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail)
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

		cache.Set("posts_cat_"+cat+"_"+page, j)
		w.Write(j)
	}
	w.Write(cached)
}

func PostsByTagHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tag := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[4])
	p, err := strconv.Atoi(page)
	if err != nil {
		log.Fatal(err)
	}

	cached, isCached := cache.Get("posts_tag_" + tag)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT 
			posts.title, 
			posts.slug, 
			posts.url, 
			posts.summary, 
			posts.date AS dt, 
			posts.sentiment, 
			COALESCE(posts.image, ""), 
			posts.category_id, cats.slug, 
			COALESCE(cats.thumbnail, "") 
			FROM aggregator_post AS posts 
			INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title 
			INNER JOIN aggregator_post_tags AS post_tags ON post_tags.post_id = posts.title 
			INNER JOIN aggregator_tags AS tags ON post_tags.tags_id = tags.title 
			WHERE tags.slug='%[1]s' 
			ORDER BY dt DESC 
			LIMIT %[2]d,%[3]d;`, tag, postsPerPage*p, postsPerPage)
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date,
				&post.Sentiment, &post.Image, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail)
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

		cache.Set("posts_tag_"+tag, j)
		w.Write(j)
	}
	w.Write(cached)
}

func PostsByCalendarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	year := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	month := url.QueryEscape(strings.Split(r.RequestURI, "/")[4])
	day := url.QueryEscape(strings.Split(r.RequestURI, "/")[5])
	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[6])
	p, err := strconv.Atoi(page)
	if err != nil {
		log.Fatal(err)
	}

	cached, isCached := cache.Get("posts_cal_" + year + "_" + month + "_" + day + "_" + page)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT 
			posts.title, 
			posts.slug, 
			posts.url, 
			posts.summary, 
			posts.date AS dt, 
			posts.sentiment, 
			COALESCE(posts.image, ""), 
			posts.category_id, cats.slug, 
			COALESCE(cats.thumbnail, "") 
			FROM aggregator_post AS posts 
			INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title 
			WHERE posts.date >= '%[1]s-%[2]s-%[3]s' AND posts.date < '%[1]s-%[2]s-%[3]s' + INTERVAL 1 DAY
			ORDER BY dt DESC 
			LIMIT %[4]d,%[5]d;`, year, month, day, postsPerPage*p, postsPerPage)

		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date,
				&post.Sentiment, &post.Image, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail)
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

		cache.Set("posts_cal_"+year+"_"+month+"_"+day+"_"+page, j)
		w.Write(j)
	}
	w.Write(cached)
}

func TodayPostsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	p, err := strconv.Atoi(page)
	if err != nil {
		log.Fatal(err)
	}

	cached, isCached := cache.Get("today_" + page)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		dateBack := time.Now().AddDate(0, 0, -2) //2 days back for "today"

		query := fmt.Sprintf(`SELECT 
			posts.title, 
			posts.slug, 
			posts.url, 
			posts.summary, 
			posts.date, 
			posts.sentiment, 
			COALESCE(posts.image, ""), 
			posts.category_id, 
			cats.slug, 
			COALESCE(cats.thumbnail, "") 
			FROM aggregator_post as posts 
			INNER JOIN aggregator_category as cats ON posts.category_id = cats.title 
			WHERE date > '%[1]s' ORDER BY date DESC 
			LIMIT %[2]d,%[3]d;`, dateBack, postsPerPage*p, postsPerPage)

		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date,
				&post.Sentiment, &post.Image, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail)
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

		cache.Set("today_"+page, j)
		w.Write(j)
	}
	w.Write(cached)
}

func FilledTagsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cnt := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	c, err := strconv.Atoi(cnt)
	if err != nil {
		log.Fatal(err)
	}

	cached, isCached := cache.Get("tags_pop")
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT 
			tags.title, 
			tags.slug, 
			COUNT(posts.title) AS cnt 
			FROM aggregator_tags AS tags 
			INNER JOIN aggregator_post_tags AS post_tags ON tags.title = post_tags.tags_id 
			INNER JOIN aggregator_post as posts ON post_tags.post_id = posts.title 
			GROUP BY tags.title 
			HAVING cnt > %[1]d 
			ORDER BY tags.title;`, c)

		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		tags := make([]models.Tag, 0)
		for rows.Next() {
			tag := models.Tag{}

			err := rows.Scan(&tag.Title, &tag.Slug, &tag.PostCnt)
			if err != nil {
				panic(err)
			}
			tags = append(tags, tag)
		}
		if err = rows.Err(); err != nil {
			panic(err)
		}

		j, err := json.Marshal(tags)
		if err != nil {
			log.Fatal(err)
		}

		cache.Set("tags_pop", j)
		w.Write(j)
	}
	w.Write(cached)
}

func TagsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	p, err := strconv.Atoi(page)
	if err != nil {
		log.Fatal(err)
	}

	cached, isCached := cache.Get("tags_" + page)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT 
			tags.title, 
			tags.slug, 
			count(posts.title) AS cnt 
			FROM aggregator_tags AS tags 
			INNER JOIN aggregator_post_tags AS post_tags ON tags.title = post_tags.tags_id 
			INNER JOIN aggregator_post as posts ON post_tags.post_id = posts.title 
			GROUP BY tags.title 
			ORDER BY tags.title 
			LIMIT %[1]d,%[2]d;`, catsPerPage*p, catsPerPage)

		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		tags := make([]models.Tag, 0)
		for rows.Next() {
			tag := models.Tag{}

			err := rows.Scan(&tag.Title, &tag.Slug, &tag.PostCnt)
			if err != nil {
				panic(err)
			}
			tags = append(tags, tag)
		}
		if err = rows.Err(); err != nil {
			panic(err)
		}

		j, err := json.Marshal(tags)
		if err != nil {
			log.Fatal(err)
		}

		cache.Set("tags_"+page, j)
		w.Write(j)
	}
	w.Write(cached)
}

func CategoriesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	p, err := strconv.Atoi(page)
	if err != nil {
		log.Fatal(err)
	}

	cached, isCached := cache.Get("cats_" + page)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT 
			cats.title, 
			cats.slug, 
			COALESCE(cats.thumbnail, ""), 
			count(posts.title) AS cnt 
			FROM aggregator_category AS cats 
			INNER JOIN aggregator_post AS posts ON posts.category_id = cats.title 
			GROUP BY cats.title 
			LIMIT %[1]d,%[2]d;`, catsPerPage*p, catsPerPage)

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

		cache.Set("cats_"+page, j)
		w.Write(j)
	}
	w.Write(cached)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	postSlug := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])

	cached, isCached := cache.Get("post_" + postSlug)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT 
			posts.title, 
			posts.slug, 
			posts.url, 
			posts.summary, 
			posts.date, 
			posts.sentiment, 
			COALESCE(posts.image, ""), 
			posts.category_id, 
			cats.slug, 
			COALESCE(cats.thumbnail, "") 
			FROM aggregator_post as posts 
			INNER JOIN aggregator_category as cats ON posts.category_id = cats.title 
			WHERE posts.slug='%s';`, postSlug)
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

func PopularHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	title := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])

	cached, isCached := cache.Get("popular_")
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT  
			tags.title, 
			tags.slug 
			FROM aggregator_tags AS tags 
			INNER JOIN aggregator_post_tags AS post_tags ON tags.title = post_tags.tags_id 
			INNER JOIN aggregator_post as posts ON post_tags.post_id = posts.title
			WHERE posts.slug='%[1]s';`, title)

		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		tags := make([]models.Tag, 0)
		for rows.Next() {
			tag := models.Tag{}

			err := rows.Scan(&tag.Title, &tag.Slug)
			if err != nil {
				panic(err)
			}
			tags = append(tags, tag)
		}
		if err = rows.Err(); err != nil {
			panic(err)
		}

		j, err := json.Marshal(tags)
		if err != nil {
			log.Fatal(err)
		}

		cache.Set("popular_", j)
		w.Write(j)

	}
	w.Write(cached)
}
