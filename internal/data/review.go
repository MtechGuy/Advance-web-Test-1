// Filename: internal/data/comments.go
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mtechguy/test1/internal/validator"
)

// each name begins with uppercase so that they are exportable/public
type Review struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"version"`
}

type ReviewModel struct {
	DB *sql.DB
}

func ValidateReview(v *validator.Validator, review *Review) {

	v.Check(review.Content != "", "content", "must be provided")
	// check if the Author field is empty
	v.Check(review.Author != "", "author", "must be provided")
	// check if the Content field is empty
	v.Check(len(review.Content) <= 100, "content", "must not be more than 100 bytes long")
	// check if the Author field is empty
	v.Check(len(review.Author) <= 25, "author", "must not be more than 25 bytes long")
}

func (c ReviewModel) InsertReview(review *Review) error {
	// the SQL query to be executed against the database table
	query := `
		 INSERT INTO comments (content, author)
		 VALUES ($1, $2)
		 RETURNING id, created_at, version
		 `
	// the actual values to replace $1, and $2
	args := []any{review.Content, review.Author}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the comments database table. We ask for the the
	// id, created_at, and version to be sent back to us which we will use
	// to update the Comment struct later on
	return c.DB.QueryRowContext(ctx, query, args...).Scan(
		&review.ID,
		&review.CreatedAt,
		&review.Version)
}

// Get a specific Comment from the comments table
func (c ReviewModel) GetReview(id int64) (*Review, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
		 SELECT id, created_at, content, author, version
		 FROM comments
		 WHERE id = $1
	   `
	// declare a variable of type Comment to store the returned comment
	var review Review

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&review.ID,
		&review.CreatedAt,
		&review.Content,
		&review.Author,
		&review.Version,
	)
	// Cont'd on the next slide
	// check for which type of error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &review, nil
}

func (c ReviewModel) UpdateReview(review *Review) error {
	// The SQL query to be executed against the database table
	// Every time we make an update, we increment the version number
	query := `
			UPDATE comments
			SET content = $1, author = $2, version = version + 1
			WHERE id = $3
			RETURNING version 
			`

	args := []any{review.Content, review.Author, review.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&review.Version)

}

func (c ReviewModel) DeleteReview(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM comments
        WHERE id = $1
		`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// ExecContext does not return any rows unlike QueryRowContext.
	// It only returns  information about the the query execution
	// such as how many rows were affected
	result, err := c.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// Probably a wrong id was provided or the client is trying to
	// delete an already deleted comment
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

func (c ReviewModel) GetAllReviews(content, author string, filters Filters) ([]*Review, Metadata, error) {
	// Construct the SQL query with placeholders for parameters
	query := fmt.Sprintf(`
	SELECT COUNT(*) OVER(), id, created_at, content, author, version
	FROM comments
	WHERE (to_tsvector('simple', content) @@ plainto_tsquery('simple', $1) OR $1 = '') 
	  AND (to_tsvector('simple', author) @@ plainto_tsquery('simple', $2) OR $2 = '') 
	ORDER BY %s %s, id ASC 
	LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	// Set a context with a 3-second timeout for query execution
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query with provided filters and parameters
	rows, err := c.DB.QueryContext(ctx, query, content, author, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var totalRecords int
	reviews := []*Review{}

	// Iterate over result rows and scan data into Review struct
	for rows.Next() {
		var review Review
		if err := rows.Scan(&totalRecords, &review.ID, &review.CreatedAt, &review.Content, &review.Author, &review.Version); err != nil {
			return nil, Metadata{}, err
		}
		reviews = append(reviews, &review)
	}

	// Check if any error occurred during row iteration
	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Calculate metadata for pagination
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return reviews, metadata, nil
}
