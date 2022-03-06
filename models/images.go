package models

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Image is not stored in the DB
type Image struct {
	GalleryID uint
	Filename string
	URL string
}

func (i *Image) Path() string {
	temp := url.URL{
		Path: "/" + i.RelativePath(),
	}
	return temp.String()
}

func (i *Image) FullPath(id uint, filename string) string {
	host := os.Getenv("HOST")
	return host + i.RelativePath()
}

func (i *Image) RelativePath() string {
	return fmt.Sprintf("images/galleries/%v/%v", i.GalleryID, i.Filename)
}

type ImageService interface {
	Create(galleryID uint, r io.ReadCloser, filename string) error
	ByGalleryID(galleryID uint) ([]Image, error)
	Delete(i *Image) error
}

func NewImageService() ImageService {
	return &imageService{}
}

type imageService struct {}

func (is *imageService) ByGalleryID(galleryID uint) ([]Image, error) {
	path := is.imagePath(galleryID)
	fileStrings, err := filepath.Glob(path + "*")
	if err != nil {
		return nil, err
	}

	ret := make([]Image, len(fileStrings))
	for i := range fileStrings {
		fileStrings[i] = strings.Replace(fileStrings[i], path, "", 1)
		ret[i] = Image{
			Filename: fileStrings[i],
			GalleryID: galleryID,
		}
	}
	return ret, nil
}

func (is *imageService) Delete(i *Image) error {
	return os.Remove(i.RelativePath())
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