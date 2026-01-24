package files

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type Uploader struct {
	UploadDir string
}

func NewUploader(dir string) *Uploader {
	_ = os.MkdirAll(dir, os.ModePerm)
	return &Uploader{UploadDir: dir}
}

func (u *Uploader) SaveFile(file *multipart.FileHeader, subDir string) (string, error) {
	src, err := file.Open()
	if err != nil { return "", err }
	defer src.Close()

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	finalPath := filepath.Join(u.UploadDir, subDir, filename)
	_ = os.MkdirAll(filepath.Dir(finalPath), os.ModePerm)

	dst, err := os.Create(finalPath)
	if err != nil { return "", err }
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil { return "", err }

	// Return the relative path to store in DB
	return filepath.Join("uploads", subDir, filename), nil
}

func (u *Uploader) DeleteFile(relativeURL string) error {
	// convert "uploads/pdf/file.pdf" to "./uploads/pdf/file.pdf"
	fullPath := filepath.Join(".", relativeURL)
	return os.Remove(fullPath)
}