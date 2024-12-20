package post

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alissoncorsair/appsolidario-backend/service/auth"
	"github.com/alissoncorsair/appsolidario-backend/service/user"
	"github.com/alissoncorsair/appsolidario-backend/storage"
	"github.com/alissoncorsair/appsolidario-backend/types"
	"github.com/alissoncorsair/appsolidario-backend/utils"
	"github.com/google/uuid"
)

type Handler struct {
	postStore *Store
	userStore *user.Store
	storage   *storage.R2Storage
}

func NewHandler(postStore *Store, userStore *user.Store, storage *storage.R2Storage) *Handler {
	return &Handler{
		postStore: postStore,
		userStore: userStore,
		storage:   storage,
	}
}

func (h *Handler) HandleCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("failed to parse form: %w", err))
		return
	}

	var payload types.CreatePostRequest
	payload.Description = r.FormValue("description")

	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get user: %w", err))
		return
	}

	post := &types.Post{
		UserID:      userID,
		AuthorName:  user.Name,
		Description: payload.Description,
	}

	if files := r.MultipartForm.File["photos"]; len(files) > 0 {
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to open file: %w", err))
				return
			}
			defer file.Close()

			_, filename, err := h.storage.UploadFile(r.Context(), file, fileHeader.Filename)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to upload file: %w", err))
				return
			}

			post.Photos = append(post.Photos, filename)
		}
	}

	createdPost, err := h.postStore.CreatePost(post)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	userPicture, err := h.userStore.GetUserProfilePicture(userID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get user profile picture: %w", err))
		return
	}

	if userPicture != nil {
		createdPost.UserPicture = userPicture.Path
	}

	utils.WriteJSON(w, http.StatusCreated, createdPost)
}

func (h *Handler) HandleDeletePost(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	if id == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid post ID"))
		return
	}

	postID, err := strconv.Atoi(id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid post ID"))
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())

	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	post, err := h.postStore.GetPostByID(postID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("post not found"))
		return
	}

	if post.UserID != userID {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("you don't have permission to delete this post"))
		return
	}

	err = h.postStore.DeletePost(postID, h.storage)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete post: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Post deleted successfully"})
}

func (h *Handler) HandleCreateComment(w http.ResponseWriter, r *http.Request) {
	var id = r.PathValue("post_id")

	if id == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid post ID"))
		return
	}

	postID, err := strconv.Atoi(id)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid post ID"))
		return
	}

	var payload types.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid request payload"))
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get user: %w", err))
		return
	}

	comment := &types.Comment{
		PostID:     postID,
		UserID:     userID,
		AuthorName: user.Name,
		Content:    payload.Content,
	}

	createdComment, err := h.postStore.CreateComment(comment)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, createdComment)
}

func (h *Handler) HandleDeleteComment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid comment ID"))
		return
	}

	commentID, err := strconv.Atoi(id)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid comment ID"))
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())

	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	comment, err := h.postStore.GetCommentByID(commentID)

	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("comment not found"))
		return
	}

	if comment.UserID != userID {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("you don't have permission to delete this comment"))
		return
	}

	err = h.postStore.DeleteComment(commentID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete comment: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Comment deleted successfully"})
}

func (h *Handler) HandleUploadPhoto(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("failed to parse form: %w", err))
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("failed to get file: %w", err))
		return
	}
	defer file.Close()

	filename := uuid.New().String() + filepath.Ext(header.Filename)

	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create upload directory: %w", err))
		return
	}

	dst, err := os.Create(filepath.Join(uploadDir, filename))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create file: %w", err))
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to save file: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"filename": filename})
}

func (h *Handler) HandleGetPhoto(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path)
	if filename == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid filename"))
		return
	}

	file, err := h.storage.GetFile(r.Context(), filename)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get file: %w", err))
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "image/jpeg")
	io.Copy(w, file)
}

func (h *Handler) HandleGetPostByID(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	if id == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid post ID"))
		return
	}

	postID, err := strconv.Atoi(id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid post ID"))
		return
	}

	post, err := h.postStore.GetPostByID(postID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("post not found"))
		return
	}

	post.UserPicture = ""
	userPicture, err := h.userStore.GetUserProfilePicture(post.UserID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get user profile picture: %w", err))
		return
	}

	if userPicture != nil {
		post.UserPicture = userPicture.Path
	}

	for _, comment := range post.Comments {
		commentProfilePicture, err := h.userStore.GetUserProfilePicture(comment.UserID)

		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get user profile picture: %w", err))
			return
		}

		if commentProfilePicture != nil {
			comment.UserPicture = commentProfilePicture.Path
		}
	}

	utils.WriteJSON(w, http.StatusOK, post)
}

func (h *Handler) HandleGetOwnPosts(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	posts, err := h.postStore.GetPostsByUserID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get posts: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, posts)
}

func (h *Handler) HandleGetPostsByUserId(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	if id == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	userID, err := strconv.Atoi(id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	posts, err := h.postStore.GetPostsByUserID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get posts: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, posts)
}

func (h *Handler) HandleGetPostsByCity(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	city := r.URL.Query().Get("city")
	if city == "" {
		user, err := h.userStore.GetUserByID(userID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get user: %w", err))
			return
		}
		city = user.City
	}

	posts, err := h.postStore.GetPostsByCity(city)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get posts: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, posts)
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /posts", auth.WithJWTAuth(h.HandleCreatePost, h.userStore))
	router.HandleFunc("GET /post/{id}", auth.WithJWTAuth(h.HandleGetPostByID, h.userStore))
	router.HandleFunc("GET /posts/user/{id}", auth.WithJWTAuth(h.HandleGetPostsByUserId, h.userStore))
	router.HandleFunc("GET /me/posts", auth.WithJWTAuth(h.HandleGetOwnPosts, h.userStore))
	router.HandleFunc("DELETE /posts/{id}", auth.WithJWTAuth(h.HandleDeletePost, h.userStore))
	router.HandleFunc("POST /comments/{post_id}", auth.WithJWTAuth(h.HandleCreateComment, h.userStore))
	router.HandleFunc("DELETE /comments/{id}", auth.WithJWTAuth(h.HandleDeleteComment, h.userStore))
	router.HandleFunc("GET /photos/{filename}", h.HandleGetPhoto)
	router.HandleFunc("GET /posts/city", auth.WithJWTAuth(h.HandleGetPostsByCity, h.userStore))
}
