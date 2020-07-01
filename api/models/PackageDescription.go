package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
)

// PackageDescription is the DB table structure and json input structure for an address assignment. It is a one to many relationship table. One user can have many addresses
type PackageDescription struct {
	ID         uint64      `gorm:"primary_key;auto_increment" json:"id"`
	Contents   null.String `gorm:"size:255;" json:"contents"`
	OrderLink  null.String `gorm:"size:255;" json:"order_link"`
	OrderImage null.String `gorm:"size:255;" json:"order_image"`
	CreatedAt  time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// Prepare formats the PackageDescription object
func (pd *PackageDescription) Prepare() {
	pd.ID = 0
	pd.CreatedAt = time.Now()
	pd.UpdatedAt = time.Now()
}

// SavePackageDescription is used to save a package. It is called when a mail carrier makes a new shipping API request
func (pd *PackageDescription) SavePackageDescription(db *gorm.DB) (*PackageDescription, error) {
	err := db.Debug().Create(&pd).Error
	if err != nil {
		return &PackageDescription{}, err
	}
	return pd, nil
}

// UpdatePackageDescription is used to update the delivered status
func (pd *PackageDescription) UpdatePackageDescription(db *gorm.DB, ID uint64, contents null.String, orderLink null.String, orderImage null.String) (*PackageDescription, error) {

	var err error
	err = db.Debug().Model(&PackageDescription{}).Where("id = ?", ID).Updates(PackageDescription{Contents: contents, OrderLink: orderLink, OrderImage: orderImage, UpdatedAt: time.Now()}).Error
	if err != nil {
		return &PackageDescription{}, err
	}
	return pd, nil
}

// FindAllPackageDescriptionsForUser retrieves the last 100 non-delivered packages (with tracking numbers) a user has linked to their account. Used in UI to provide users currently active packages
func (pd *PackageDescription) FindAllPackageDescriptionsForUser(db *gorm.DB, uid uuid.UUID) (*[]PackageDescription, error) {
	var err error
	packages := []PackageDescription{}
	err = db.Debug().Set("gorm:auto_preload", true).Model(&PackageDescription{}).Where("user_id = ?", uid).Limit(100).Find(&packages).Error
	if err != nil {
		return &[]PackageDescription{}, err
	}
	return &packages, nil
}

// DeletePackageDescription removes an address assignment from the DB (should never use this unless correcting an accidental addition)
func (pd *PackageDescription) DeletePackageDescription(db *gorm.DB, did uint64) (int64, error) {

	db = db.Debug().Model(&PackageDescription{}).Where("id = ?", did).Take(&PackageDescription{}).Delete(&PackageDescription{})

	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("PackageDescription not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
