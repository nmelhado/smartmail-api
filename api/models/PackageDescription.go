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
	ID          uint64      `gorm:"primary_key;auto_increment" json:"id"`
	Package     Package     `json:"package"`
	PackageID   uint64      `gorm:"not null;" sql:"type:int REFERENCES packages(id)" json:"package_id"`
	User        User        `json:"user"`
	UserID      uuid.UUID   `gorm:"type:uuid; not null;" sql:"type:uuid REFERENCES users(id)" json:"user_id"`
	Description null.String `gorm:"size:255;" json:"description"`
	CreatedAt   time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// Prepare formats the PackageDescription object
func (pd *PackageDescription) Prepare() {
	pd.ID = 0
	pd.Package = Package{}
	pd.User = User{}
	pd.CreatedAt = time.Now()
	pd.UpdatedAt = time.Now()
}

// SavePackageDescription is used to save a package. It is called when a mail carrier makes a new shipping API request
func (pd *PackageDescription) SavePackageDescription(db *gorm.DB) error {
	newPackageDescription := PackageDescription{}
	err := db.Debug().Model(&PackageDescription{}).Where("user_id = ? AND package_id = ?", pd.UserID, pd.PackageID).Attrs(PackageDescription{PackageID: pd.PackageID, UserID: pd.UserID, Description: pd.Description, CreatedAt: time.Now()}).FirstOrCreate(&newPackageDescription).Error
	if err != nil {
		return err
	}
	return nil
}

// UpdatePackageDescription is used to update the delivered status
func (pd *PackageDescription) UpdatePackageDescription(db *gorm.DB, ID uint64, description null.String) (*PackageDescription, error) {

	var err error
	err = db.Debug().Model(&PackageDescription{}).Where("id = ?", ID).Updates(PackageDescription{Description: description, UpdatedAt: time.Now()}).Error
	if err != nil {
		return &PackageDescription{}, err
	}
	return pd, nil
}

// FindAllPackageDescriptionsForUser retrieves the last 100 non-delivered packages (with tracking numbers) a user has linked to their account. Used in UI to provide users currently active packages
func (pd *PackageDescription) FindAllPackageDescriptionsForUser(db *gorm.DB, uid uuid.UUID) (*[]PackageDescription, error) {
	var err error
	packages := []PackageDescription{}
	err = db.Debug().Set("gorm:auto_preload", true).Order("created_at asc").Model(&PackageDescription{}).Where("user_id = ?", uid).Limit(100).Find(&packages).Error
	if err != nil {
		return &[]PackageDescription{}, err
	}
	return &packages, nil
}

// FindAndSortAllPackageDescriptionsForUser retrieves the last 100 packages (with tracking numbers) a user has linked to their account. Used in UI to provide packages status and history
// Sorts results into delivered and open
func (pd *PackageDescription) FindAndSortAllPackageDescriptionsForUser(db *gorm.DB, uid uuid.UUID) (*[]PackageDescription, *[]PackageDescription, error) {
	openPackageDescriptions := []PackageDescription{}
	deliveredPackageDescriptions := []PackageDescription{}

	packageDescriptions, err := pd.FindAllPackageDescriptionsForUser(db, uid)
	if err != nil {
		return &[]PackageDescription{}, &[]PackageDescription{}, err
	}

	for _, packageDescription := range *packageDescriptions {
		if packageDescription.Package.Delivered {
			deliveredPackageDescriptions = append(deliveredPackageDescriptions, packageDescription)
		} else {
			openPackageDescriptions = append(openPackageDescriptions, packageDescription)
		}
	}

	return &openPackageDescriptions, &deliveredPackageDescriptions, nil
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
