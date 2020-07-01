package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
)

// Package is the DB table structure and json input structure for an address assignment. It is a one to many relationship table. One user can have many addresses
type Package struct {
	ID                   uint64             `gorm:"primary_key;auto_increment" json:"id"`
	MailCarrier          APIUser            `json:"mail_carrier"`
	MailCarrierID        uuid.UUID          `gorm:"type:uuid;not null;" sql:"type:uuid REFERENCES mail_carriers(id)" json:"mail_carrier_id"`
	Sender               User               `json:"sender"`
	SenderID             uuid.NullUUID      `gorm:"type:uuid;" sql:"type:uuid REFERENCES users(id)" json:"sender_id"`
	Recipient            User               `json:"recipient"`
	RecipientID          uuid.NullUUID      `gorm:"type:uuid;" sql:"type:uuid REFERENCES users(id)" json:"recipient_id"`
	Tracking             null.String        `gorm:"size:255;" json:"tracking"`
	PackageDescription   PackageDescription `json:"package_description"`
	PackageDescriptionID uint64             `sql:"type:bigint REFERENCES package_descriptions(id)" json:"package_description_id"`
	EstimatedDelivery    null.Time          `gorm:"default:null" json:"estimated_delivery"`
	Delivered            bool               `json:"delivered"`
	DeliveredOn          null.Time          `gorm:"default:null" json:"delivered_on"`
	CreatedAt            time.Time          `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time          `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// Prepare formats the Package object
func (p *Package) Prepare() {
	p.ID = 0
	p.MailCarrier = APIUser{}
	p.Sender = User{}
	p.Recipient = User{}
	p.Delivered = false
	p.DeliveredOn = null.TimeFromPtr(nil)
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
}

// SavePackage is used to save a package. It is called when a mail carrier makes a new shipping API request
func (p *Package) SavePackage(db *gorm.DB) error {
	newPackage := Package{}
	err := db.Debug().Model(&Package{}).Where("sender_id = ? AND recipient_id = ? AND tracking = ? AND tracking IS NOT NULL", p.SenderID, p.RecipientID, p.Tracking).Attrs(Package{MailCarrierID: p.MailCarrierID, SenderID: p.SenderID, RecipientID: p.RecipientID, Tracking: p.Tracking, PackageDescriptionID: p.PackageDescriptionID, CreatedAt: time.Now()}).FirstOrCreate(&newPackage).Error
	if err != nil {
		return err
	}
	return nil
}

// UpdatePackage is used to update the delivered status
func (p *Package) UpdatePackage(db *gorm.DB, uid uuid.UUID, tracking string, delivered bool, deliveredOn null.Time, estimatedDelivery null.Time) (*Package, error) {
	var err error
	err = db.Debug().Model(&Package{}).Where("tracking = ? AND (sender_id = ? OR recipient_id = ?)", tracking, uid, uid).Updates(Package{Delivered: delivered, DeliveredOn: deliveredOn, EstimatedDelivery: estimatedDelivery, UpdatedAt: time.Now()}).Error
	if err != nil {
		return &Package{}, err
	}
	return p, nil
}

// SetPackageDescription is used to set the package description
func (p *Package) SetPackageDescription(db *gorm.DB, packageID uint64, packageDescriptionID uint64) error {
	var err error
	err = db.Debug().Model(&Package{}).Where("id = ?", packageID).Updates(Package{PackageDescriptionID: packageDescriptionID, UpdatedAt: time.Now()}).Error
	if err != nil {
		return err
	}
	return nil
}

// FindPackageByTrackingAndShipper retrieves the data for a package by using the sender ID and the tracking code
func (p *Package) FindPackageByTrackingAndShipper(db *gorm.DB, senderID uuid.UUID, tracking string) (*Package, error) {
	var err error
	err = db.Debug().Model(Package{}).Where("sender_id = ? AND tracking = ?", senderID, tracking).Take(&p).Error
	if err != nil {
		return &Package{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &Package{}, errors.New("User Not Found")
	}
	return p, err
}

// FindAllOpenPackagesForUser retrieves the last 100 non-delivered packages (with tracking numbers) a user has linked to their account. Used in UI to provide users currently active packages
func (p *Package) FindAllOpenPackagesForUser(db *gorm.DB, uid uuid.UUID) (*[]Package, error) {
	var err error
	packages := []Package{}
	err = db.Debug().Set("gorm:auto_preload", true).Order("delivered_on desc, estimated_delivery desc").Model(&Package{}).Where("(sender_id = ? OR recipient_id = ?) AND tracking IS NOT NULL AND tracking <> '' AND delivered = false", uid, uid).Limit(100).Find(&packages).Error
	if err != nil {
		return &[]Package{}, err
	}
	return &packages, nil
}

// FindAllDeliveredPackagesForUser retrieves the last 100 delivered packages (with tracking numbers) a user has linked to their account. Used in UI to provide users with package history
func (p *Package) FindAllDeliveredPackagesForUser(db *gorm.DB, uid uuid.UUID) (*[]Package, error) {
	var err error
	packages := []Package{}
	err = db.Debug().Set("gorm:auto_preload", true).Order("delivered_on desc, estimated_delivery desc").Model(&Package{}).Where("(sender_id = ? OR recipient_id = ?) AND tracking IS NOT NULL AND tracking <> '' AND delivered = true", uid, uid).Limit(100).Find(&packages).Error
	if err != nil {
		return &[]Package{}, err
	}
	return &packages, nil
}

// FindAllPackagesForUser retrieves the last 100 non-delivered and last 100 delivered packages (with tracking numbers) a user has linked to their account. Used in UI to provide packages status and history
// Uses the FindAllOpenPackagesForUser and FindAllDeliveredPackagesForUser functions to retrieve the results
func (p *Package) FindAllPackagesForUser(db *gorm.DB, uid uuid.UUID) (*[]Package, *[]Package, error) {
	openPackages, err := p.FindAllOpenPackagesForUser(db, uid)
	if err != nil {
		return &[]Package{}, &[]Package{}, err
	}
	deliveredPackages, err := p.FindAllDeliveredPackagesForUser(db, uid)
	if err != nil {
		return &[]Package{}, &[]Package{}, err
	}
	return openPackages, deliveredPackages, nil
}

// DeletePackage removes an address assignment from the DB (should never use this unless correcting an accidental addition)
func (p *Package) DeletePackage(db *gorm.DB, did uint64) (int64, error) {

	db = db.Debug().Model(&Package{}).Where("id = ?", did).Take(&Package{}).Delete(&Package{})

	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Package not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
