package models

import "gorm.io/gorm"

type Gallery struct {
	gorm.Model
	UserId uint `gorm:"not_null;index"`
	Title string `gorm:"not_null"`
}

type GalleryService interface {
	GalleryDB
}

type galleryValidator struct {
	GalleryDB
}



func (gv *galleryValidator) Create(gallery *Gallery) error {
	// Order of functions passed in to validator is important!
	err := runGalleryValFuncs(
		gallery, gv.titleRequired, gv.userIDRequired,
		); 
	if err != nil {
		return err
	}
	return gv.GalleryDB.Create(gallery)
}

func NewGalleryService(db *gorm.DB) GalleryService {
	return &galleryService{
		GalleryDB: &galleryValidator{
			&galleryGorm{db}},
	}
}

type galleryService struct {
	GalleryDB
}

var _ GalleryDB = &galleryGorm{}

type GalleryDB interface{
	Create(gallery *Gallery) error 
	ByID (id uint) (*Gallery, error)
}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
		return gg.db.Create(gallery).Error
}

func (gg *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery
	db := gg.db.Where("id = ?", id)
	err := first(db, &gallery)
	return &gallery, err
}

func (gv *galleryValidator) userIDRequired(g *Gallery) error {
	if g.UserId <= 0 {
		return ErrUserIDRequired
	}
	return nil
}

func (gv *galleryValidator) titleRequired(g *Gallery) error {
	if g.Title == ""  {
		return ErrTitleRequired
	}
	return nil
}

type galleryValFunc func(*Gallery) error

func runGalleryValFuncs(gallery *Gallery, fns ...galleryValFunc) error {
	for _, fn := range fns {
		if err := fn(gallery); err != nil {
			return err
		}
	}
	return nil
}