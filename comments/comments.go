package comments

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Comment struct {
	CommentID int       `json:"comment_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Comment   string    `json:"comment"`
	PostID    int       `json:"post_id"`
}

func FetchAllComments(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		postIdStr := c.Params("post_id")
		fmt.Println("postIdStr:", postIdStr)

		// Convert postIdStr to an integer
		postId, err := strconv.Atoi(postIdStr)
		if err != nil {
			fmt.Println("error converting postIdStr to integer:", err)
			return err
		}

		// Fetch all comments for the specified post
		rows, err := db.Query("SELECT comment_id, user_id, created_at, comment FROM comments WHERE post_id = $1", postId)
		if err != nil {
			fmt.Println("error fetching comments ", err)
			return err
		}
		defer rows.Close()

		var comments []Comment
		for rows.Next() {
			var comment Comment
			err := rows.Scan(&comment.CommentID, &comment.UserID, &comment.CreatedAt, &comment.Comment)
			if err != nil {
				fmt.Println("error scanning into comment", err)
				return err
			}
			comments = append(comments, comment)
		}
		// Return success response with comments data
		return c.Render("showComments", fiber.Map{
			"Comments": comments,
		}, "layouts/empty") // Return an empty layout to prevent rendering of the entire layout
	}
}

// func to post comments
func PostComment(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request body into Comment object
		var newComment Comment

		if err := c.BodyParser(&newComment); err != nil {
			fmt.Println("failed to pass into body")
			return err
		}

		if newComment.Comment == "" {
			fmt.Println("comment is enpty")
			return errors.New("empty comment provided")

		}

		// Set the creation time
		newComment.CreatedAt = time.Now()

		// Execute SQL statement to insert the comment into the database
		_, err := db.Exec("INSERT INTO comments (post_id, user_id, comment, created_at) VALUES($1, $2, $3, $4)",
			newComment.PostID, newComment.UserID, newComment.Comment, newComment.CreatedAt)
		if err != nil {
			fmt.Println("failed to insert into the db", err)
			return err
		}

		// Return success response
		return c.Render("comment", fiber.Map{
			"Comment": newComment,
		}, "layouts/empty")

	}
}
