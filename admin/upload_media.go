package admin

import (
	"mime/multipart"
	"path/filepath"
	"strings"
)

const (
	maxImageUploadSize = 10 << 20  // 10MB
	maxVideoUploadSize = 100 << 20 // 100MB
	maxDocUploadSize   = 50 << 20  // 50MB
)

func classifyUploadFile(file *multipart.FileHeader) (kind string, maxSize int64, ok bool) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	ct := strings.ToLower(file.Header.Get("Content-Type"))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".heic":
		return "image", maxImageUploadSize, true
	case ".mp4", ".mov", ".webm", ".m4v", ".avi", ".mkv":
		return "video", maxVideoUploadSize, true
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".zip":
		return "doc", maxDocUploadSize, true
	}
	if strings.HasPrefix(ct, "image/") {
		return "image", maxImageUploadSize, true
	}
	if strings.HasPrefix(ct, "video/") {
		return "video", maxVideoUploadSize, true
	}
	if strings.HasPrefix(ct, "application/pdf") {
		return "doc", maxDocUploadSize, true
	}
	return "", 0, false
}
