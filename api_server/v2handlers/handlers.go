package v2handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/die-net/lrucache"
	"github.com/xenu256/qprob_goapi/api_server/database"
	"github.com/xenu256/qprob_goapi/api_server/models"
)

var cache = lrucache.New(104857600*3, 60*60*24) //300 Mb, 24 hours
var postsPerPage = 20
var catsPerPage = 40

func TagsByPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	title := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	if len(title) == 0 {
		return
	}

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
			return
		}
		defer rows.Close()

		tags := make([]models.Tag, 0)
		for rows.Next() {
			tag := models.Tag{}

			err := rows.Scan(&tag.Title, &tag.Slug)
			if err != nil {
				return
			}
			tags = append(tags, tag)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(tags)
		if err != nil {
			return
		}

		cache.Set("post_tags_"+title, j)
		w.Write(j)
		w.WriteHeader(http.StatusOK)

	}
	w.Write(cached)
	w.WriteHeader(http.StatusOK)
}

func PostsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	p, err := strconv.Atoi(page)
	if err != nil {
		return
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
			COALESCE(posts.wordcloud, ""), 
			posts.category_id, 
			cats.slug, 
			COALESCE(cats.thumbnail, ""), 
			posts.hits, 
			(SELECT 
				COUNT(*) 
				FROM aggregator_post) 
			FROM aggregator_post as posts
			INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title 
			ORDER BY posts.date DESC 
			LIMIT %[1]d,%[2]d;`, postsPerPage*p, postsPerPage)

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date,
				&post.Sentiment, &post.Image, &post.Wordcloud, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail, &post.Hits, &post.TotalPosts)
			if err != nil {
				return
			}
			posts = append(posts, post)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(posts)
		if err != nil {
			return
		}

		cache.Set("posts_"+page, j)
		w.Write(j)
	}
	w.Write(cached)
}

// experimental benchmark improvement, test needs disabling the json cache and MySQL cache
func PostsHandler2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	p, err := strconv.Atoi(page)
	if err != nil {
		return
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
			COALESCE(cats.thumbnail, ""), 
			posts.hits 
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
			return
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date,
				&post.Sentiment, &post.Image, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail, &post.Hits)
			if err != nil {
				return
			}
			posts = append(posts, post)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(posts)
		if err != nil {
			return
		}

		cache.Set("posts2_"+page, j)
		w.Write(j)
	}
	w.Write(cached)
}

func PostsByCatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cat := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	if len(cat) == 0 {
		return
	}
	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[4])
	p, err := strconv.Atoi(page)
	if err != nil {
		return
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
			CASE posts.dead 
				WHEN 0 THEN "" 
				WHEN 1 THEN posts.content 
			END AS content, 
			posts.date AS dt, 
			posts.sentiment, 
			COALESCE(posts.image, ""), 
			COALESCE(posts.wordcloud, ""), 
			posts.category_id, 
			cats.slug, 
			COALESCE(cats.thumbnail, ""), 
			posts.hits, 
			(SELECT 
				COUNT(*) 
				FROM aggregator_post AS posts 
				INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title 
				WHERE cats.slug='%[1]s'), 
			posts.dead 
			FROM aggregator_post AS posts 
			INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title 
			WHERE cats.slug='%[1]s' 
			ORDER BY dt DESC 
			LIMIT %[2]d,%[3]d;`, cat, postsPerPage*p, postsPerPage)
		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Content, &post.Date,
				&post.Sentiment, &post.Image, &post.Wordcloud, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail, &post.Hits, &post.TotalPosts, &post.Status)
			if err != nil {
				return
			}
			posts = append(posts, post)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(posts)
		if err != nil {
			return
		}

		cache.Set("posts_cat_"+cat+"_"+page, j)
		w.Write(j)
	}
	w.Write(cached)
}

func PostsByTagHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tag := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	if len(tag) == 0 {
		return
	}
	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[4])
	p, err := strconv.Atoi(page)
	if err != nil {
		return
	}

	cached, isCached := cache.Get("posts_tag_" + tag + "_" + page)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT 
			posts.title, 
			posts.slug, 
			posts.url, 
			posts.summary, 
			CASE posts.dead 
				WHEN 0 THEN "" 
				WHEN 1 THEN posts.content 
			END AS content, 
			posts.date AS dt, 
			posts.sentiment, 
			COALESCE(posts.image, ""), 
			COALESCE(posts.wordcloud, ""), 
			posts.category_id, cats.slug, 
			COALESCE(cats.thumbnail, ""), 
			posts.hits, 
			(SELECT 
				COUNT(*) 
				FROM aggregator_post AS posts 
				INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title 
				INNER JOIN aggregator_post_tags AS post_tags ON post_tags.post_id = posts.title 
				INNER JOIN aggregator_tags AS tags ON post_tags.tags_id = tags.title 
				WHERE tags.slug='%[1]s'), 
			posts.dead 
			FROM aggregator_post AS posts 
			INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title 
			INNER JOIN aggregator_post_tags AS post_tags ON post_tags.post_id = posts.title 
			INNER JOIN aggregator_tags AS tags ON post_tags.tags_id = tags.title 
			WHERE tags.slug='%[1]s' 
			ORDER BY dt DESC 
			LIMIT %[2]d,%[3]d;`, tag, postsPerPage*p, postsPerPage)
		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Content, &post.Date,
				&post.Sentiment, &post.Image, &post.Wordcloud, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail, &post.Hits, &post.TotalPosts, &post.Status)
			if err != nil {
				return
			}
			posts = append(posts, post)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(posts)
		if err != nil {
			return
		}

		cache.Set("posts_tag_"+tag+"_"+page, j)
		w.Write(j)
	}
	w.Write(cached)
}

func PostsByCalendarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	year := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	month := url.QueryEscape(strings.Split(r.RequestURI, "/")[4])
	day := url.QueryEscape(strings.Split(r.RequestURI, "/")[5])
	if len(year) != 4 || len(month) != 2 || len(day) != 2 {
		return
	}
	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[6])
	p, err := strconv.Atoi(page)
	if err != nil {
		return
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
			COALESCE(posts.wordcloud, ""), 
			posts.category_id, cats.slug, 
			COALESCE(cats.thumbnail, ""), 
			posts.hits, 
			(SELECT 
				COUNT(*) 
				FROM FROM aggregator_post AS posts
				INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title 
				WHERE posts.date >= '%[1]s-%[2]s-%[3]s' AND posts.date < '%[1]s-%[2]s-%[3]s' + INTERVAL 1 DAY) 
			FROM aggregator_post AS posts 
			INNER JOIN aggregator_category AS cats ON posts.category_id = cats.title 
			WHERE posts.date >= '%[1]s-%[2]s-%[3]s' AND posts.date < '%[1]s-%[2]s-%[3]s' + INTERVAL 1 DAY
			ORDER BY dt DESC 
			LIMIT %[4]d,%[5]d;`, year, month, day, postsPerPage*p, postsPerPage)

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date,
				&post.Sentiment, &post.Image, &post.Wordcloud, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail, &post.Hits, &post.TotalPosts)
			if err != nil {
				return
			}
			posts = append(posts, post)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(posts)
		if err != nil {
			return
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
		return
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
			CASE posts.dead 
				WHEN 0 THEN "" 
				WHEN 1 THEN posts.content 
			END AS content, 
			posts.date, 
			posts.sentiment, 
			COALESCE(posts.image, ""), 
			COALESCE(posts.wordcloud, ""), 
			posts.category_id, 
			cats.slug, 
			COALESCE(cats.thumbnail, ""), 
			posts.hits, 
			(SELECT 
				COUNT(*) 
				FROM aggregator_post as posts
				INNER JOIN aggregator_category as cats ON posts.category_id = cats.title 
				WHERE date > '%[1]s'), 
			posts.dead 
			FROM aggregator_post as posts 
			INNER JOIN aggregator_category as cats ON posts.category_id = cats.title 
			WHERE date > '%[1]s' 
			ORDER BY date DESC 
			LIMIT %[2]d,%[3]d;`, dateBack, postsPerPage*p, postsPerPage)

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Content, &post.Date,
				&post.Sentiment, &post.Image, &post.Wordcloud, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail, &post.Hits, &post.TotalPosts, &post.Status)
			if err != nil {
				return
			}
			posts = append(posts, post)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(posts)
		if err != nil {
			return
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
		return
	}

	cached, isCached := cache.Get("tags_pop_" + cnt)
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
			ORDER BY tags.title 
			LIMIT 100;`, c)

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		tags := make([]models.Tag, 0)
		for rows.Next() {
			tag := models.Tag{}

			err := rows.Scan(&tag.Title, &tag.Slug, &tag.PostCnt)
			if err != nil {
				return
			}
			tags = append(tags, tag)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(tags)
		if err != nil {
			return
		}

		cache.Set("tags_pop_"+cnt, j)
		w.Write(j)
	}
	w.Write(cached)
}

func TopTagsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cached, isCached := cache.Get("top_tags_")
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := `SELECT 
			tags.title, 
			tags.slug, 
			COUNT(posts.title) AS cnt 
			FROM aggregator_tags AS tags 
			INNER JOIN aggregator_post_tags AS post_tags ON tags.title = post_tags.tags_id 
			INNER JOIN aggregator_post as posts ON post_tags.post_id = posts.title 
			GROUP BY tags.title 
			ORDER BY COUNT(*) DESC 
			LIMIT 40;`

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		tags := make([]models.Tag, 0)
		for rows.Next() {
			tag := models.Tag{}

			err := rows.Scan(&tag.Title, &tag.Slug, &tag.PostCnt)
			if err != nil {
				return
			}
			tags = append(tags, tag)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(tags)
		if err != nil {
			return
		}

		cache.Set("top_tags_", j)
		w.Write(j)
	}
	w.Write(cached)
}

func TagsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	p, err := strconv.Atoi(page)
	if err != nil {
		return
	}

	cached, isCached := cache.Get("tags_" + page)
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
			ORDER BY tags.title 
			LIMIT %[1]d,%[2]d;`, catsPerPage*p, catsPerPage)

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		tags := make([]models.Tag, 0)
		for rows.Next() {
			tag := models.Tag{}

			err := rows.Scan(&tag.Title, &tag.Slug, &tag.PostCnt)
			if err != nil {
				return
			}
			tags = append(tags, tag)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(tags)
		if err != nil {
			return
		}

		cache.Set("tags_"+page, j)
		w.Write(j)
	}
	w.Write(cached)
}

func TopCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cached, isCached := cache.Get("top_cats_")
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := `SELECT 
			cats.title, 
			cats.slug, 
			COALESCE(cats.thumbnail, ""), 
			COUNT(posts.title) AS cnt 
			FROM aggregator_category AS cats 
			INNER JOIN aggregator_post AS posts ON posts.category_id = cats.title 
			GROUP BY cats.title 
			ORDER BY COUNT(*) DESC 
			LIMIT 21;`

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		categories := make([]models.Category, 0)
		for rows.Next() {
			category := models.Category{}

			err := rows.Scan(&category.Title, &category.Slug, &category.Thumbnail,
				&category.PostCnt)
			if err != nil {
				return
			}

			categories = append(categories, category)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(categories)
		if err != nil {
			return
		}

		cache.Set("top_cats_", j)
		w.Write(j)
	}
	w.Write(cached)
}

func CategoriesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	p, err := strconv.Atoi(page)
	if err != nil {
		return
	}

	cached, isCached := cache.Get("cats_" + page)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT 
			cats.title, 
			cats.slug, 
			COALESCE(cats.thumbnail, ""), 
			COUNT(posts.title) AS cnt, 
			(SELECT 
				COUNT(*) 
				FROM aggregator_category) 
			FROM aggregator_category AS cats 
			INNER JOIN aggregator_post AS posts ON posts.category_id = cats.title 
			GROUP BY cats.title 
			LIMIT %[1]d,%[2]d;`, catsPerPage*p, catsPerPage)

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		categories := make([]models.Category, 0)
		for rows.Next() {
			category := models.Category{}

			err := rows.Scan(&category.Title, &category.Slug, &category.Thumbnail,
				&category.PostCnt, &category.TotalCats)
			if err != nil {
				return
			}

			categories = append(categories, category)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(categories)
		if err != nil {
			return
		}

		cache.Set("cats_"+page, j)
		w.Write(j)
	}
	w.Write(cached)
}

func AllCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cached, isCached := cache.Get("allcats_")
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := `SELECT 
			cats.title, 
			cats.slug, 
			COALESCE(cats.thumbnail, ""), 
			COUNT(posts.title) AS cnt 
			FROM aggregator_category AS cats 
			INNER JOIN aggregator_post AS posts ON posts.category_id = cats.title 
			GROUP BY cats.title;`

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		categories := make([]models.Category, 0)
		for rows.Next() {
			category := models.Category{}

			err := rows.Scan(&category.Title, &category.Slug, &category.Thumbnail,
				&category.PostCnt)
			if err != nil {
				return
			}

			categories = append(categories, category)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(categories)
		if err != nil {
			return
		}

		cache.Set("allcats_", j)
		w.Write(j)
	}
	w.Write(cached)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	postSlug := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	if len(postSlug) == 0 {
		return
	}

	cached, isCached := cache.Get("post_" + postSlug)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT 
			posts.title, 
			posts.slug, 
			posts.url, 
			posts.summary, 
			CASE posts.dead 
				WHEN 0 THEN "" 
				WHEN 1 THEN posts.content 
			END AS content, 
			posts.date, 
			posts.sentiment, 
			COALESCE(posts.image, ""), 
			COALESCE(posts.wordcloud, ""), 
			posts.category_id, 
			cats.slug, 
			COALESCE(cats.thumbnail, ""), 
			posts.hits, 
			posts.dead 
			FROM aggregator_post as posts 
			INNER JOIN aggregator_category as cats ON posts.category_id = cats.title 
			WHERE posts.slug='%s';`, postSlug)
		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		query = fmt.Sprintf(`UPDATE aggregator_post 
			SET hits = hits + 1 
			WHERE slug='%[1]s';`, postSlug)

		r, err := db.Exec(query)
		if err != nil {
			return
		}
		count, err := r.RowsAffected()
		if err != nil || count != 1 {
			return
		}

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary,
				&post.Content, &post.Date, &post.Sentiment, &post.Image, &post.Wordcloud,
				&post.CategoryID.Title, &post.CategoryID.Slug, &post.CategoryID.Thumbnail,
				&post.Hits, &post.Status)
			if err != nil {
				return
			}
			posts = append(posts, post)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(posts)
		if err != nil {
			return
		}

		cache.Set("post_"+postSlug, j)
		w.Write(j)
	}
	w.Write(cached)
}

func PopularPostsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	hits := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	h, err := strconv.Atoi(hits)
	if err != nil {
		return
	}
	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[4])
	p, err := strconv.Atoi(page)
	if err != nil {
		return
	}

	cached, isCached := cache.Get("popular_" + hits + "_" + page)
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
			COALESCE(posts.wordcloud, ""), 
			posts.category_id, 
			cats.slug, 
			COALESCE(cats.thumbnail, ""), 
			posts.hits, 
			(SELECT 
				COUNT(*) 
				FROM aggregator_post as posts
				INNER JOIN aggregator_category as cats ON posts.category_id = cats.title
				WHERE hits > %[1]d) 
			FROM aggregator_post as posts 
			INNER JOIN aggregator_category as cats ON posts.category_id = cats.title
			WHERE hits > %[1]d
			LIMIT %[2]d,%[3]d;`, h, postsPerPage*p, postsPerPage)

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date,
				&post.Sentiment, &post.Image, &post.Wordcloud, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail, &post.Hits, &post.TotalPosts)
			if err != nil {
				return
			}
			posts = append(posts, post)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(posts)
		if err != nil {
			return
		}

		cache.Set("popular_"+hits+"_"+page, j)
		w.Write(j)

	}
	w.Write(cached)
}

func PostsByPopularityHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	p, err := strconv.Atoi(page)
	if err != nil || p > 20 {
		return
	}

	cached, isCached := cache.Get("popular_" + page)
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
			COALESCE(posts.wordcloud, ""), 
			posts.category_id, 
			cats.slug, 
			COALESCE(cats.thumbnail, ""), 
			posts.hits, 
			(SELECT 
				COUNT(*) 
				FROM aggregator_post
				WHERE hits > 10), 
			posts.dead 
			FROM aggregator_post as posts 
			INNER JOIN aggregator_category as cats ON posts.category_id = cats.title 
			WHERE posts.hits > 10 
			ORDER BY hits DESC 
			LIMIT %[1]d,%[2]d;`, postsPerPage*p, postsPerPage)

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Date,
				&post.Sentiment, &post.Image, &post.Wordcloud, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail, &post.Hits, &post.TotalPosts, &post.Status)
			if err != nil {
				return
			}
			posts = append(posts, post)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(posts)
		if err != nil {
			return
		}

		cache.Set("popular_"+page, j)
		w.Write(j)

	}
	w.Write(cached)
}

func MostPopularPostsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	p, err := strconv.Atoi(page)
	if err != nil || p > 20 {
		return
	}

	cached, isCached := cache.Get("most_popular_" + page)
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := fmt.Sprintf(`SELECT 
			posts.title, 
			posts.slug, 
			posts.url, 
			posts.summary, 
			CASE posts.dead 
				WHEN 0 THEN "" 
				WHEN 1 THEN posts.content 
			END AS content, 
			posts.date, 
			posts.sentiment, 
			COALESCE(posts.image, ""), 
			COALESCE(posts.wordcloud, ""), 
			posts.category_id, 
			cats.slug, 
			COALESCE(cats.thumbnail, ""), 
			posts.hits, 
			(SELECT 
				COUNT(*) 
				FROM aggregator_post
				WHERE YEAR(posts.date) = YEAR(CURRENT_DATE()) 
				AND MONTH(posts.date) = MONTH(CURRENT_DATE())
				AND hits > 10), 
			posts.dead 
			FROM aggregator_post as posts 
			INNER JOIN aggregator_category as cats ON posts.category_id = cats.title
			WHERE YEAR(posts.date) = YEAR(CURRENT_DATE()) 
			AND MONTH(posts.date) = MONTH(CURRENT_DATE()) 
			AND posts.hits > 10 
			ORDER BY hits DESC 
			LIMIT %[1]d,%[2]d;`, postsPerPage*p, postsPerPage)

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		posts := make([]models.Post, 0)
		for rows.Next() {
			post := models.Post{}
			err := rows.Scan(&post.Title, &post.Slug, &post.URL, &post.Summary, &post.Content, &post.Date,
				&post.Sentiment, &post.Image, &post.Wordcloud, &post.CategoryID.Title, &post.CategoryID.Slug,
				&post.CategoryID.Thumbnail, &post.Hits, &post.TotalPosts, &post.Status)
			if err != nil {
				return
			}
			posts = append(posts, post)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(posts)
		if err != nil {
			return
		}

		cache.Set("most_popular_"+page, j)
		w.Write(j)

	}
	w.Write(cached)
}

func UpdatePostHitHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	postSlug := url.QueryEscape(strings.Split(r.RequestURI, "/")[3])
	if len(postSlug) == 0 {
		return
	}

	db := database.Connect()
	defer db.Close()

	query := fmt.Sprintf(`UPDATE aggregator_post 
		SET hits = hits + 1 
		WHERE slug='%[1]s';`, postSlug)

	rows, err := db.Exec(query)
	if err != nil {
		return
	}
	count, err := rows.RowsAffected()
	if err != nil || count != 1 {
		w.Write([]byte(`{"status": 400}`))
	} else {
		w.Write([]byte(`{"status": 200}`))
	}

}

func MetaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cached, isCached := cache.Get("meta_")
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := `SELECT 
			(SELECT 
				COUNT(*) 
				FROM aggregator_category) AS cats, 
			(SELECT 
				COUNT(*) 
				FROM aggregator_post) AS posts, 
			(SELECT 
				COUNT(*) 
				FROM aggregator_tags) AS tags;`

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		meta := make([]models.Meta, 0)
		for rows.Next() {
			m := models.Meta{}

			err := rows.Scan(&m.CatTotal, &m.PostTotal, &m.TagTotal)
			if err != nil {
				return
			}

			meta = append(meta, m)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(meta)
		if err != nil {
			return
		}

		cache.Set("meta_", j)
		w.Write(j)
	}
	w.Write(cached)
}

func SentimentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cached, isCached := cache.Get("sentiment_")
	if isCached == false {
		db := database.Connect()
		defer db.Close()

		query := `SELECT 
			date, 
			COALESCE(sentiment, 0.0) 
			FROM aggregator_post 
			WHERE date > DATE_FORMAT(DATE_SUB(CURRENT_DATE, INTERVAL 1 MONTH), '%Y-%m-%d') 
			ORDER BY date DESC;`

		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		sents := make([]models.Sentiment, 0)
		for rows.Next() {
			s := models.Sentiment{}
			err := rows.Scan(&s.Date, &s.Sentiment)
			if err != nil {
				return
			}
			sents = append(sents, s)
		}
		if err = rows.Err(); err != nil {
			return
		}

		j, err := json.Marshal(sents)
		if err != nil {
			return
		}

		cache.Set("sentiment_", j)
		w.Write(j)

	}
	w.Write(cached)
}
