package models

import "gorm.io/gorm"

type Gallery struct {
	gorm.Model
	UserId uint `gorm:"not_null;index"`
	Title string `gorm:"not_null"`
	Images []Image `gorm:"-"`
	UserToken string `gorm:"-"`
}

// Used to 
func (g *Gallery) ImagesSplitN(n int) [][]Image {
	ret := make([][]Image, n)
	for i := 0; i < n; i++ {
		ret[i] = make([]Image, 0)
	}
	for i, img := range g.Images {
		bucket := i % n
		ret[bucket] = append(ret[bucket], img)
	}
	return ret
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

func (gv *galleryValidator) Update(gallery *Gallery) error {
	// Order of functions passed in to validator is important!
	err := runGalleryValFuncs(
		gallery, gv.titleRequired, gv.userIDRequired,
		); 
	if err != nil {
		return err
	}
	return gv.GalleryDB.Update(gallery)
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
	Update(gallery *Gallery) error
	ByID (id uint) (*Gallery, error)
	ByUserID(UserID uint) ([]Gallery, error)
	Delete (id uint) error
}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
		return gg.db.Create(gallery).Error
}

func (gg *galleryGorm) Update(gallery *Gallery) error {
	return gg.db.Save(gallery).Error
}

func (gg *galleryGorm) Delete(id uint) error {
	gallery := Gallery{Model: gorm.Model{ID:id}}
	return gg.db.Delete(&gallery).Error
}

func (gg *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery
	db := gg.db.Where("id = ?", id)
	err := first(db, &gallery)
	return &gallery, err
}

func (gg *galleryGorm) ByUserID(userID uint) ([]Gallery, error) {
	var galleries []Gallery
	err := gg.db.Where("user_id = ?", userID).Find(&galleries).Error
	if err != nil {
		return nil, err
	}
	return galleries, nil
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

func (gv *galleryValidator) Delete(id uint) error {
	var gallery Gallery
	gallery.ID = id
	if id <= 0 {
		return ErrIDInvalid
	}

	return gv.GalleryDB.Delete(id)
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