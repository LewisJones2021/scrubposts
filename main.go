package main

import (
	"database/sql"
	"fmt"
	"log"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"

	_ "github.com/lib/pq"
)

// data object (struct) that represent the Post fields.
type Post struct {
	ID                int       `json:"id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	PhotoURL          string    `json:"photo_url"`
	PhotoURLAfter     string    `json:"photo_url_after"`
	DateCreated       time.Time `json:"date_created"`
	DisplayTime       string    `json:"display_time"`
	Hashtags          string    `json:"hashtags"`
	SelectedHashtag   string    `json:"selectedHashtags"`
	AvailableHashtags string    `json:"availableHashtags"`
	Likes             int       `json:"likes"`
}

// map to store number of likes for each post
var likeCounts map[int]int

func main() {
	// open a connection to the data base
	db, err := sql.Open("postgres", "user=postgres dbname=scrubposts sslmode=disable")
	if err != nil {
		log.Fatal("Error launching db", err)
	}
	defer db.Close()

	engine := html.New("./templates", ".html")
	engine.Reload(true)

	// create a new fiber instance with the configuration.
	app := fiber.New(
		fiber.Config{
			Views:                 engine,
			ViewsLayout:           "layouts/main", // add this to config
			PassLocalsToViews:     true,
			DisableStartupMessage: false,
		},
	)

	// serve static files
	app.Static("/", "./public")
	// serve uploads
	app.Static("/uploads", "./uploads")

	// get endpoints

	// define the route for the HTMX homepage
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("postForm", fiber.Map{"PhotoURL": ""})
	})

	// api endpoint for the viewpost page.
	app.Get("/viewPost", func(c *fiber.Ctx) error {

		// Get the selected hashtag from the query parameters
		selectedHashtag := c.Query("hashtags")
		fmt.Println("fetched hashtags", selectedHashtag)
		query := "SELECT id, title, description, photo_url, photo_url_after, hashtags, date_created, likes FROM posts "
		fmt.Println(query)

		// If a hashtag is provided, modify the query to filter posts by the hashtag
		if selectedHashtag != "" {
			query += "WHERE hashtags LIKE '%" + selectedHashtag + "%'"
		}
		query += "ORDER BY date_created DESC"

		// fetch data from the db
		rows, err := db.Query(query)
		fmt.Println("fetched")
		if err != nil {
			fmt.Println("Error: can't select queries from the db", err)
			return err
		}
		defer rows.Close()

		// scan the db columns
		var posts []Post
		for rows.Next() {
			var post Post
			if err := rows.Scan(&post.ID, &post.Title, &post.Description, &post.PhotoURL, &post.PhotoURLAfter, &post.Hashtags, &post.DateCreated, &post.Likes); err != nil {
				fmt.Println("Error: error scanning the data", err)
				return err
			}
			posts = append(posts, post)
		}

		// loop around the posts array, dispaly time for each post in the array
		for i := range posts {
			posts[i].DisplayTime = posts[i].DateCreated.Format("02/01/2006 at 15:05:05")
			fmt.Println(posts[i].DisplayTime)
		}

		// Set up a slice of available hashtags for the form
		availableHashtags := []string{"bathrooms", "kitchens", "windows", "ovens"}

		// Return a success message
		return c.Render("viewPost", fiber.Map{
			"Posts":             posts,
			"SelectedHashtag":   selectedHashtag,
			"AvailableHashtags": availableHashtags,
		})
	})

	// post endpoints

	// Initialize like counts map
	likeCounts = make(map[int]int)
	app.Post("/like/:id", func(c *fiber.Ctx) error {

		// Parse post ID from URL params
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			fmt.Println("Error getting the id from the url params", err)
		}

		// Increment like count for the post
		likeCounts[id]++

		// Increment like count for the post in the database
		_, err = db.Exec("UPDATE posts SET likes = likes +1 WHERE id = $1", id)
		if err != nil {
			fmt.Println("Error updating the likes in the db:", err)
			return err
		}

		// Return updated like count from the database
		var likes int
		err = db.QueryRow("SELECT likes FROM posts WHERE id = $1", id).Scan(&likes)
		if err != nil {
			fmt.Println("Error fetching like count from database:", err)
			return err
		}
		// Return updated like count
		return c.SendString(strconv.Itoa(likes))
	})

	// Define an API endpoint for storing posts
	app.Post("/posts", func(c *fiber.Ctx) error {
		// Parse the JSON request body into a Post struct
		var post Post
		if err := c.BodyParser(&post); err != nil {
			return err
		}
		fmt.Println("posted")

		// store in db here
		post.DateCreated = time.Now().UTC()
		// Insert post into the PostgreSQL database
		_, err := db.Exec("INSERT INTO POSTS (title, description, photo_url, photo_url_after, hashtags, date_created) VALUES($1, $2, $3, $4, $5, $6)",
			post.Title, post.Description, post.PhotoURL, post.PhotoURLAfter, post.Hashtags, post.DateCreated)
		if err != nil {
			fmt.Println("Error instering into post table, line 90:", err)
			return err
		}

		// Return a success message
		return c.Render("successTemplate", fiber.Map{
			"Post": post,
		})
	})

	// post endpoint to upload image
	app.Post("/upload", func(c *fiber.Ctx) error {
		// Parse the form data, including file uploads
		form, err := c.MultipartForm()
		if err != nil {
			fmt.Println("error parsing form:", err)
			return err
		}
		defer form.RemoveAll()

		// Process each uploaded file
		var uploadedFiles []string
		for _, files := range [][]*multipart.FileHeader{form.File["file"], form.File["afterFile"]} {
			// Save the file to the server
			for _, file := range files {
				err := c.SaveFile(file, "./uploads/"+file.Filename)

				if err != nil {
					fmt.Println("error saving file:", err)
					return err
				}
				uploadedFiles = append(uploadedFiles, "/uploads/"+file.Filename)
			}
		}
		// Check if no files were uploaded
		if len(uploadedFiles) == 0 {
			fmt.Println("no files uploaded")
			return c.SendStatus(fiber.StatusBadRequest)
		}

		// Return the URLs of the uploaded files
		return c.Render("postForm", fiber.Map{"PhotoURL": uploadedFiles[0], "PhotoURLAfter": uploadedFiles[1]})
	})

	// run the server
	if err := app.Listen(":3000"); err != nil {
		fmt.Println(err)
	}

	fmt.Println("Server running")
}
