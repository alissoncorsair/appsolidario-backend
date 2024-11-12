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
        INSERT INTO posts (user_id, author_name, description)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at
    `
	err = tx.QueryRow(query, post.UserID, post.AuthorName, post.Description).
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

	post.Comments = []*types.Comment{}

	if len(post.Photos) == 0 {
		post.Photos = []string{}
	}

	return post, nil
}

func (s *Store) GetPostByID(id int) (*types.Post, error) {
	query := `
        SELECT p.id, p.user_id, p.author_name, p.description, p.created_at, p.updated_at
        FROM posts p
        WHERE p.id = $1
    `
	var post types.Post
	err := s.db.QueryRow(query, id).Scan(
		&post.ID, &post.UserID, &post.AuthorName, &post.Description,
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

	comments, err := s.GetCommentsByPostID(id)

	if err != nil {
		return nil, fmt.Errorf("error getting comments: %w", err)
	}

	post.Comments = comments

	return &post, nil
}

func (s *Store) GetPostsByCity(city string) ([]*types.Post, error) {
	query := `
        SELECT p.id, p.user_id, p.description, p.author_name, p.created_at, p.updated_at, pp.path
        FROM posts p
        JOIN users u ON p.user_id = u.id
        LEFT JOIN profile_pictures pp ON u.id = pp.user_id
        WHERE u.city = $1
        ORDER BY p.created_at DESC
    `
	rows, err := s.db.Query(query, city)

	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}
	defer rows.Close()

	var posts []*types.Post
	for rows.Next() {
		var post types.Post
		var userPicture sql.NullString
		if err := rows.Scan(&post.ID, &post.UserID, &post.Description, &post.AuthorName, &post.CreatedAt, &post.UpdatedAt, &userPicture); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		if userPicture.Valid {
			post.UserPicture = userPicture.String
		}

		comments, err := s.GetCommentsByPostID(post.ID)

		if err != nil {
			return nil, fmt.Errorf("error getting comments: %w", err)
		}

		post.Comments = comments

		if len(post.Comments) == 0 {
			post.Comments = []*types.Comment{}
		}

		photos, err := s.GetPhotosByPostID(post.ID)

		if err != nil {
			return nil, fmt.Errorf("error getting photos: %w", err)
		}

		for _, photo := range photos {
			post.Photos = append(post.Photos, photo.Filename)
		}

		if len(post.Photos) == 0 {
			post.Photos = []string{}
		}

		posts = append(posts, &post)
	}

	if len(posts) == 0 {
		return []*types.Post{}, nil
	}

	return posts, nil
}

func (s *Store) GetPhotosByPostID(postID int) ([]types.PostPhoto, error) {
	query := `
        SELECT id, post_id, filename, created_at
        FROM post_photos
        WHERE post_id = $1
        ORDER BY id ASC
    `
	rows, err := s.db.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get photos: %w", err)
	}
	defer rows.Close()

	var photos []types.PostPhoto
	for rows.Next() {
		var photo types.PostPhoto
		if err := rows.Scan(&photo.ID, &photo.PostID, &photo.Filename, &photo.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan photo: %w", err)
		}
		photos = append(photos, photo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over photos: %w", err)
	}

	if len(photos) == 0 {
		return []types.PostPhoto{}, nil
	}

	return photos, nil
}

func (s *Store) GetPostsByUserID(id int) ([]*types.Post, error) {

	query := `
	SELECT p.id, p.user_id, p.author_name, p.description, p.created_at, p.updated_at, pp.path
	FROM posts p
	JOIN users u ON p.user_id = u.id
	LEFT JOIN profile_pictures pp ON u.id = pp.user_id
	WHERE p.user_id = $1
	ORDER BY p.created_at DESC
`

	rows, err := s.db.Query(query, id)

	if err != nil {
		return nil, fmt.Errorf("error getting posts: %w", err)
	}

	defer rows.Close()

	var posts []*types.Post
	for rows.Next() {
		var post types.Post
		var userPicture sql.NullString
		if err := rows.Scan(
			&post.ID, &post.UserID, &post.AuthorName, &post.Description,
			&post.CreatedAt, &post.UpdatedAt, &userPicture,
		); err != nil {
			return nil, fmt.Errorf("error scanning post: %w", err)
		}

		if userPicture.Valid {
			post.UserPicture = userPicture.String
		}

		comments, err := s.GetCommentsByPostID(post.ID)

		if err != nil {
			return nil, fmt.Errorf("error getting comments: %w", err)
		}

		post.Comments = comments

		if len(post.Comments) == 0 {
			post.Comments = []*types.Comment{}
		}

		photos, err := s.GetPhotosByPostID(post.ID)

		if err != nil {
			return nil, fmt.Errorf("error getting photos: %w", err)
		}

		for _, photo := range photos {
			post.Photos = append(post.Photos, photo.Filename)
		}

		posts = append(posts, &post)
	}

	return posts, nil
}

func (s *Store) CreateComment(comment *types.Comment) (*types.Comment, error) {
	query := `
        INSERT INTO comments (post_id, user_id, author_name, content)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at
    `
	err := s.db.QueryRow(query, comment.PostID, comment.UserID, comment.AuthorName, comment.Content).
		Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("error creating comment: %w", err)
	}

	return comment, nil
}

func (s *Store) GetCommentByID(commentID int) (*types.Comment, error) {
	query := `
		SELECT id, post_id, user_id, author_name, content, created_at, updated_at
		FROM comments
		WHERE id = $1
	`
	var comment types.Comment
	err := s.db.QueryRow(query, commentID).Scan(
		&comment.ID, &comment.PostID, &comment.UserID, &comment.AuthorName, &comment.Content,
		&comment.CreatedAt, &comment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("error getting comment: %w", err)
	}

	return &comment, nil
}

func (s *Store) DeleteComment(commentID int) error {
	result, err := s.db.Exec("DELETE FROM comments WHERE id = $1", commentID)
	if err != nil {
		return fmt.Errorf("error deleting comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

func (s *Store) GetCommentsByPostID(postID int) ([]*types.Comment, error) {
	query := `
	SELECT c.id, c.post_id, c.user_id, c.author_name, c.content, c.created_at, c.updated_at, pp.path
	FROM comments c
	JOIN users u ON c.user_id = u.id
	LEFT JOIN profile_pictures pp ON u.id = pp.user_id
	WHERE c.post_id = $1
	ORDER BY c.created_at ASC
`
	rows, err := s.db.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("error getting comments: %w", err)
	}
	defer rows.Close()

	var comments []*types.Comment
	for rows.Next() {
		var comment types.Comment
		var userPicture sql.NullString
		if err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &comment.AuthorName, &comment.Content,
			&comment.CreatedAt, &comment.UpdatedAt, &userPicture,
		); err != nil {
			return nil, fmt.Errorf("error scanning comment: %w", err)
		}

		if userPicture.Valid {
			comment.UserPicture = userPicture.String
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
