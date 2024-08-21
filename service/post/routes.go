package post

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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
	payload.Title = r.FormValue("title")
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

	post := &types.Post{
		UserID:      userID,
		Title:       payload.Title,
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

	utils.WriteJSON(w, http.StatusCreated, createdPost)
}

func (h *Handler) HandleCreateComment(w http.ResponseWriter, r *http.Request) {
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

	comment := &types.Comment{
		PostID:  payload.PostID,
		UserID:  userID,
		Content: payload.Content,
	}

	createdComment, err := h.postStore.CreateComment(comment)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, createdComment)
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

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /posts", auth.WithJWTAuth(h.HandleCreatePost, h.userStore))
	router.HandleFunc("POST /comments", auth.WithJWTAuth(h.HandleCreateComment, h.userStore))
	router.HandleFunc("GET /photos/{filename}", h.HandleGetPhoto)
}
