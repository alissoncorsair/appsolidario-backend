package post

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alissoncorsair/appsolidario-backend/storage"
	"github.com/alissoncorsair/appsolidario-backend/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) CreatePost(post *types.Post) (*types.Post, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
        INSERT INTO posts (user_id, title, description)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at
    `
	err = tx.QueryRow(query, post.UserID, post.Title, post.Description).
		Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("error creating post: %w", err)
	}

	if len(post.Photos) > 0 {
		photoQuery := `
            INSERT INTO post_photos (post_id, filename)
            VALUES ($1, $2)
        `
		for _, filename := range post.Photos {
			_, err = tx.Exec(photoQuery, post.ID, filename)
			if err != nil {
				return nil, fmt.Errorf("error adding photo: %w", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return post, nil
}

func (s *Store) GetPostByID(id int) (*types.Post, error) {
	query := `
        SELECT p.id, p.user_id, p.title, p.description, p.created_at, p.updated_at
        FROM posts p
        WHERE p.id = $1
    `
	var post types.Post
	err := s.db.QueryRow(query, id).Scan(
		&post.ID, &post.UserID, &post.Title, &post.Description,
		&post.CreatedAt, &post.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("error getting post: %w", err)
	}

	photoQuery := `
        SELECT filename FROM post_photos
        WHERE post_id = $1
        ORDER BY created_at
    `
	rows, err := s.db.Query(photoQuery, id)
	if err != nil {
		return nil, fmt.Errorf("error getting post photos: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, fmt.Errorf("error scanning filename: %w", err)
		}
		post.Photos = append(post.Photos, filename)
	}

	return &post, nil
}

func (s *Store) CreateComment(comment *types.Comment) (*types.Comment, error) {
	query := `
        INSERT INTO comments (post_id, user_id, content)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at
    `
	err := s.db.QueryRow(query, comment.PostID, comment.UserID, comment.Content).
		Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("error creating comment: %w", err)
	}

	return comment, nil
}

func (s *Store) GetCommentsByPostID(postID int) ([]*types.Comment, error) {
	query := `
        SELECT id, post_id, user_id, content, created_at, updated_at
        FROM comments
        WHERE post_id = $1
        ORDER BY created_at DESC
    `
	rows, err := s.db.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("error getting comments: %w", err)
	}
	defer rows.Close()

	var comments []*types.Comment
	for rows.Next() {
		var comment types.Comment
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &comment.Content,
			&comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning comment: %w", err)
		}
		comments = append(comments, &comment)
	}

	return comments, nil
}

func (s *Store) DeletePost(postID int, storageClient *storage.R2Storage) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	var filenames []string
	photoQuery := "SELECT filename FROM post_photos WHERE post_id = $1"
	rows, err := tx.Query(photoQuery, postID)
	if err != nil {
		return fmt.Errorf("error fetching post photos: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return fmt.Errorf("error scanning filename: %w", err)
		}
		filenames = append(filenames, filename)
	}

	// Delete files from R2/S3
	for _, filename := range filenames {
		if err := storageClient.DeleteFile(context.Background(), filename); err != nil {
			return fmt.Errorf("error deleting file from storage: %w", err)
		}
	}

	// Delete associated comments
	_, err = tx.Exec("DELETE FROM comments WHERE post_id = $1", postID)
	if err != nil {
		return fmt.Errorf("error deleting comments: %w", err)
	}

	// Delete associated photos from database
	_, err = tx.Exec("DELETE FROM post_photos WHERE post_id = $1", postID)
	if err != nil {
		return fmt.Errorf("error deleting post photos: %w", err)
	}

	// Delete the post
	result, err := tx.Exec("DELETE FROM posts WHERE id = $1", postID)
	if err != nil {
		return fmt.Errorf("error deleting post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
