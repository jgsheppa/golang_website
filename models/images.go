package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ImageService interface {
	Create(galleryID uint, r io.ReadCloser, filename string) error
	ByGalleryID(galleryID uint) ([]string, error)
}

func NewImageService() ImageService {
	return &imageService{}
}

type imageService struct {}

func (is *imageService) ByGalleryID(galleryID uint) ([]string, error) {
	path := is.imagePath(galleryID)
	strings, err := filepath.Glob(path + "*")
	if err != nil {
		return nil, err
	}
	for i := range strings {
		strings[i] = "/" + strings[i]
	}
	return strings, nil
}

func (is *imageService) Create(galleryID uint, r io.ReadCloser, filename string) error {
	defer r.Close()

	path, err := is.makeImagePath(galleryID)
	if err != nil {
		return err
	}

		dst, err := os.Create(path + filename)
		if err != nil {
			return err
		}
		defer dst.Close()

		_, err = io.Copy(dst, r)
		if err != nil {
			return err
		}
	return nil
}

func (is *imageService) imagePath(galleryID uint) string {
	return fmt.Sprintf("images/galleries/%v/", galleryID)
}

func (is *imageService) makeImagePath(galleryID uint) (string, error) {
		// Create directory for user images
		galleryPath := fmt.Sprintf("images/galleries/%v/", galleryID)
		if err := os.MkdirAll(galleryPath, 0755); err != nil {
			return "", err
		}
		return galleryPath, nil
}