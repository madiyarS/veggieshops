package handlers

import (
	"bytes"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/utils"
)

// AdminUploadHandler сохраняет файлы в Dir и отдаёт публичный URL /uploads/…
type AdminUploadHandler struct {
	Dir string
}

func NewAdminUploadHandler(dir string) *AdminUploadHandler {
	return &AdminUploadHandler{Dir: dir}
}

// UploadProductImage POST multipart, поле file — JPEG/PNG/WebP до ~5 МБ.
func (h *AdminUploadHandler) UploadProductImage(c *gin.Context) {
	if h.Dir == "" {
		c.JSON(http.StatusNotImplemented, utils.ErrorResponse{Success: false, Error: "Загрузка не настроена (UPLOAD_DIR)"})
		return
	}
	if err := c.Request.ParseMultipartForm(8 << 20); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные формы"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Файл не передан (поле file)"})
		return
	}
	if file.Size > 5<<20 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Файл больше 5 МБ"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Допустимы .jpg, .png, .webp"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Не удалось прочитать файл"})
		return
	}
	defer f.Close()
	head := make([]byte, 512)
	n, err := f.Read(head)
	if err != nil && n == 0 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Пустой файл"})
		return
	}
	ct := http.DetectContentType(head[:n])
	okType := strings.HasPrefix(ct, "image/jpeg") || strings.HasPrefix(ct, "image/png") || ct == "image/webp"
	if !okType && ext == ".webp" && n >= 12 && bytes.HasPrefix(head[:n], []byte("RIFF")) && bytes.Equal(head[8:12], []byte("WEBP")) {
		okType = true
	}
	if !okType {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Файл не похож на изображение"})
		return
	}
	name := uuid.New().String() + ext
	dst := filepath.Join(h.Dir, name)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: "Не удалось сохранить файл"})
		return
	}
	url := "/uploads/" + name
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: gin.H{"url": url}})
}
