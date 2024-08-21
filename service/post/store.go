package post

import (
	"database/sql"
	"fmt"

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
            INSERT INTO post_photos (post_id, photo_url)
            VALUES ($1, $2)
        `
		for _, photoURL := range post.Photos {
			_, err = tx.Exec(photoQuery, post.ID, photoURL)
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
        SELECT photo_url FROM post_photos
        WHERE post_id = $1
        ORDER BY created_at
    `
	rows, err := s.db.Query(photoQuery, id)
	if err != nil {
		return nil, fmt.Errorf("error getting post photos: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var photoURL string
		if err := rows.Scan(&photoURL); err != nil {
			return nil, fmt.Errorf("error scanning photo URL: %w", err)
		}
		post.Photos = append(post.Photos, photoURL)
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
