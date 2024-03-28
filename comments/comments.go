package comments

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Comment struct {
	CommentID int       `json:"comment_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at`
	Comment   string    `json:"comment"`
	PostID    int       `json:"post_id"`
}

// FetchAllComments fetches all comments from the database
func FetchAllComments(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT comment_id, user_id, created_at, comment FROM comments")
		if err != nil {
			return err
		}
		defer rows.Close()
		var comments []Comment
		for rows.Next() {
			var comment Comment
			err := rows.Scan(&comment.CommentID, &comment.UserID, &comment.CreatedAt, &comment.Comment)
			if err != nil {
				return err
			}
			comments = append(comments, comment)

			if err := rows.Err(); err != nil {
				return err

			}

		}
		// Return success response
		return c.Render("comment", fiber.Map{
			"Comment": comments,
		})

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
