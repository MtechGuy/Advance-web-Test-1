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
type Product struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"version"`
}

type ProductModel struct {
	DB *sql.DB
}

func ValidateProduct(v *validator.Validator, product *Product) {

	v.Check(product.Content != "", "content", "must be provided")
	// check if the Author field is empty
	v.Check(product.Author != "", "author", "must be provided")
	// check if the Content field is empty
	v.Check(len(product.Content) <= 100, "content", "must not be more than 100 bytes long")
	// check if the Author field is empty
	v.Check(len(product.Author) <= 25, "author", "must not be more than 25 bytes long")
}

func (c ProductModel) InsertProduct(product *Product) error {
	// the SQL query to be executed against the database table
	query := `
		 INSERT INTO comments (content, author)
		 VALUES ($1, $2)
		 RETURNING id, created_at, version
		 `
	// the actual values to replace $1, and $2
	args := []any{product.Content, product.Author}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the comments database table. We ask for the the
	// id, created_at, and version to be sent back to us which we will use
	// to update the Comment struct later on
	return c.DB.QueryRowContext(ctx, query, args...).Scan(
		&product.ID,
		&product.CreatedAt,
		&product.Version)
}

// Get a specific Comment from the comments table
func (c ProductModel) GetProduct(id int64) (*Product, error) {
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
	var product Product

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.CreatedAt,
		&product.Content,
		&product.Author,
		&product.Version,
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
	return &product, nil
}

func (c ProductModel) UpdateProduct(product *Product) error {
	// The SQL query to be executed against the database table
	// Every time we make an update, we increment the version number
	query := `
			UPDATE comments
			SET content = $1, author = $2, version = version + 1
			WHERE id = $3
			RETURNING version 
			`

	args := []any{product.Content, product.Author, product.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&product.Version)

}

func (c ProductModel) DeleteProduct(id int64) error {

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

func (c ProductModel) GetAllProducts(content string, author string, filters Filters) ([]*Product, Metadata, error) {

	// the SQL query to be executed against the database table
	query := fmt.Sprintf(`
	SELECT COUNT(*) OVER(), id, created_at, content, author, version
	FROM comments
	WHERE (to_tsvector('simple', content) @@
		  plainto_tsquery('simple', $1) OR $1 = '') 
	AND (to_tsvector('simple', author) @@ 
		 plainto_tsquery('simple', $2) OR $2 = '') 
	ORDER BY %s %s, id ASC 
	LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, content, author, filters.limit(), filters.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	// clean up the memory that was used
	defer rows.Close()
	totalRecords := 0
	// we will store the address of each comment in our slice
	products := []*Product{}

	// process each row that is in rows

	for rows.Next() {
		var product Product
		err := rows.Scan(&totalRecords,
			&product.ID,
			&product.CreatedAt,
			&product.Content,
			&product.Author,
			&product.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// add the row to our slice
		products = append(products, &product)
	} // end of for loop

	// after we exit the loop we need to check if it generated any errors
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return products, metadata, nil

}
