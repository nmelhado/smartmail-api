package models

import (
	"database/sql/driver"
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
)

// Status is an enum type used for AddressAssignment (values cannot be added wothot altering the DB first!)
// refer to link for `Status` field: https://github.com/jinzhu/gorm/issues/1978
type Status string

/*
postgres command to create enum:
CREATE TYPE status AS ENUM (
	'permanent',
	'temporary',
	'package_only_permanent',
	'package_only_temporary',
	'mail_only_permanent',
	'mail_only_temporary',
	'expired',
	'deleted');
*/

const (
	Permanent            Status = "permanent"
	Temporary            Status = "temporary"
	PackageOnlyPermanent Status = "package_only_permanent"
	PackageOnlyTemporary Status = "package_only_temporary"
	MailOnlyPermanent    Status = "mail_only_permanent"
	MailOnlyTemporary    Status = "mail_only_temporary"
	Expired              Status = "expired"
	Deleted              Status = "deleted"
)

// Scan - not quite sure what this does
func (s *Status) Scan(value interface{}) error {
	*s = Status(value.([]byte))
	return nil
}

// Value returns the value for the Status enum type
func (s Status) Value() (driver.Value, error) {
	return string(s), nil
}

// AddressAssignment is the DB table structure and json input structure for an address assignment. It is a one to many relationship table. One user can have many addresses
type AddressAssignment struct {
	ID        uint64    `gorm:"primary_key;auto_increment" json:"id"`
	User      User      `json:"user"`
	UserID    uuid.UUID `gorm:"type:uuid" sql:"type:uuid REFERENCES users(id)" json:"user_id"`
	Address   Address   `json:"address"`
	AddressID uint64    `sql:"type:int REFERENCES addresses(id)" json:"address_id"`
	Status    Status    `sql:"type:status" json:"status"`
	StartDate time.Time `gorm:"default:CURRENT_TIMESTAMP;not null;" json:"start_date"`
	EndDate   null.Time `gorm:"default:null" json:"end_date"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

var validPackageStatus []Status = []Status{
	Permanent,
	Temporary,
	PackageOnlyPermanent,
	PackageOnlyTemporary,
}

var validMailStatus []Status = []Status{
	Permanent,
	Temporary,
	MailOnlyPermanent,
	MailOnlyTemporary,
}

var temporaryStatus []Status = []Status{
	Temporary,
	PackageOnlyTemporary,
	MailOnlyTemporary,
}

var permanentStatus []Status = []Status{
	Permanent,
	MailOnlyPermanent,
	PackageOnlyPermanent,
}

var expiredAndDeleted []Status = []Status{
	Expired,
	Deleted,
}

func contains(arr []Status, status Status) bool {
	for _, a := range arr {
		if a == status {
			return true
		}
	}
	return false
}

// Prepare formats the AddressAssignment object
func (aa *AddressAssignment) Prepare() {
	aa.ID = 0
	aa.User = User{}
	aa.Address = Address{}
	aa.CreatedAt = time.Now()
	aa.UpdatedAt = time.Now()
}

// Validate checks the input fields for an AddressAssignment to make sure everything is correct
func (aa *AddressAssignment) Validate() error {
	if status, err := aa.Status.Value(); status == "" || err != nil {
		return errors.New("Status required")
	}
	if aa.StartDate.IsZero() {
		return errors.New("Start date required")
	}
	if contains(temporaryStatus, aa.Status) {
		if !aa.EndDate.Valid {
			return errors.New("End date required for temporary address")
		}
	}
	return nil
}

// SaveAddressAssignment is used to save an address assignment. It is called once a user already exists and the address has been created
func (aa *AddressAssignment) SaveAddressAssignment(db *gorm.DB) (*AddressAssignment, error) {
	var err error
	if contains(temporaryStatus, aa.Status) {
		conflictingAddresses := []AddressAssignment{}
		err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND status IN ('temporary', ?) AND ((start_date <= ? AND start_date >= ?) OR (end_date >= ? AND end_date <= ?))", aa.UserID, aa.Status, aa.EndDate, aa.StartDate, aa.StartDate, aa.EndDate).Limit(1).Find(&conflictingAddresses).Error
		if err != nil {
			return &AddressAssignment{}, err
		}
		if len(conflictingAddresses) > 0 {
			return &AddressAssignment{}, errors.New("There is a conflict with another temporary address change - please make sure that the dates for temporary addresses don't overlap")
		}
	}
	err = db.Debug().Model(&AddressAssignment{}).Create(&aa).Error
	if err != nil {
		return &AddressAssignment{}, err
	}
	if aa.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", aa.UserID).Take(&aa.User).Error
		if err != nil {
			return &AddressAssignment{}, err
		}
		err = db.Debug().Model(&Address{}).Where("id = ?", aa.AddressID).Take(&aa.Address).Error
		if err != nil {
			return &AddressAssignment{}, err
		}
		if contains(permanentStatus, aa.Status) {
			err = db.Debug().Model(&AddressAssignment{}).Where("status IN (?) AND id <> ? AND user_id = ? AND end_date IS NULL", permanentStatus, aa.ID, aa.UserID).Updates(AddressAssignment{EndDate: null.TimeFrom(aa.StartDate.AddDate(0, 0, -1)), UpdatedAt: time.Now()}).Error
		}
	}
	return aa, nil
}

// UpdateAddress removes an address assignment from the DB (should never use this unless correcting an accidental addition)
func (aa *AddressAssignment) UpdateAddress(db *gorm.DB, aaid uint64, originalStart time.Time) error {

	err := db.Debug().Model(&AddressAssignment{}).Where("id = ?", aa.ID).Updates(AddressAssignment{StartDate: aa.StartDate, EndDate: aa.EndDate, UpdatedAt: time.Now()}).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("Address not found")
		}
		return err
	}

	if aa.Status == Permanent && aa.StartDate.Format("2006-01-02") != originalStart.Format("2006-01-02") {
		priorAddress := AddressAssignment{}
		err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND status = ? AND end_date = ?", aa.UserID, Permanent, originalStart.AddDate(0, 0, -1)).Find(&priorAddress).Error
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return errors.New("Could not find previous address")
			}
			return err
		}
		err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND id = ?", aa.UserID, priorAddress.ID).Updates(AddressAssignment{EndDate: null.TimeFrom(aa.StartDate.AddDate(0, 0, -1)), UpdatedAt: time.Now()}).Error
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return errors.New("Address not found")
			}
			return db.Error
		}

	}

	return nil
}

// UpdateAddressAssignment is used to update status as well as start and end dates.
func (aa *AddressAssignment) UpdateAddressAssignment(db *gorm.DB) (*AddressAssignment, error) {

	var err error
	err = db.Debug().Model(&AddressAssignment{}).Where("id = ?", aa.ID).Updates(AddressAssignment{Status: aa.Status, StartDate: aa.StartDate, EndDate: null.TimeFrom(aa.EndDate.Time.AddDate(0, 0, -1)), UpdatedAt: time.Now()}).Error
	if err != nil {
		return &AddressAssignment{}, err
	}
	return aa, nil
}

// FindMailingAddressWithSmartID allows a mailcarrier to retrieve the correct address to send mail to a user by inputing a User (retieved through SmartID) and an estimated date of delivery
func (aa *AddressAssignment) FindMailingAddressWithSmartID(db *gorm.DB, user User, targetDate time.Time) (*AddressAssignment, error) {
	var err error
	address := AddressAssignment{}

	err = db.Debug().Set("gorm:auto_preload", true).Model(&AddressAssignment{}).Where("user_id = ? AND status IN (?, ?) AND start_date < ? AND (end_date IS NULL OR end_date > ?)", user.ID, MailOnlyTemporary, Temporary, targetDate, targetDate).Find(&address).Error
	if err != nil {
		err = db.Debug().Set("gorm:auto_preload", true).Model(&AddressAssignment{}).Where("user_id = ? AND status IN (?) AND start_date < ? AND (end_date IS NULL OR end_date > ?)", user.ID, validMailStatus, targetDate, targetDate).Find(&address).Error
		if err != nil {
			return &AddressAssignment{}, err
		}
	}
	return &address, nil
}

// FindPackageAddressWithSmartID allows a mailcarrier to retrieve the correct address to send packages to a user by inputing a User (retieved through SmartID) and an estimated date of delivery
func (aa *AddressAssignment) FindPackageAddressWithSmartID(db *gorm.DB, user User, targetDate time.Time) (*AddressAssignment, error) {
	var err error
	address := AddressAssignment{}

	err = db.Debug().Set("gorm:auto_preload", true).Model(&AddressAssignment{}).Where("user_id = ? AND status IN (?, ?) AND start_date < ? AND (end_date IS NULL OR end_date > ?)", user.ID, PackageOnlyTemporary, Temporary, targetDate, targetDate).Find(&address).Error
	if err != nil {
		err = db.Debug().Set("gorm:auto_preload", true).Model(&AddressAssignment{}).Where("user_id = ? AND status IN (?) AND start_date < ? AND (end_date IS NULL OR end_date > ?)", user.ID, validPackageStatus, targetDate, targetDate).Find(&address).Error
		if err != nil {
			return &AddressAssignment{}, err
		}
	}
	return &address, nil
}

// FindAllActiveAddressesForUser retrieves the last 100 active addresses a user has linked to their account. Used in UI to provide users currently active addresses
func (aa *AddressAssignment) FindAllActiveAddressesForUser(db *gorm.DB, uid uuid.UUID) (*[]AddressAssignment, error) {
	var err error
	addresses := []AddressAssignment{}
	today := strings.Split(time.Now().String(), " ")[0]
	err = db.Debug().Set("gorm:auto_preload", true).Model(&AddressAssignment{}).Where("user_id = ? AND status NOT IN (?) AND (end_date IS NULL OR end_date > ?)", uid, expiredAndDeleted, today).Limit(100).Find(&addresses).Error
	if err != nil {
		return &[]AddressAssignment{}, err
	}
	return &addresses, nil
}

// FindAllAddressesForUser retrieves the last 100 addresses a user has linked to their account. Used in UI to provide address history
func (aa *AddressAssignment) FindAllAddressesForUser(db *gorm.DB, uid uuid.UUID) (*[]AddressAssignment, error) {
	var err error
	addresses := []AddressAssignment{}
	err = db.Debug().Set("gorm:auto_preload", true).Model(&AddressAssignment{}).Where("user_id = ? AND status <> ?", uid, Deleted).Limit(100).Find(&addresses).Error
	if err != nil {
		return &[]AddressAssignment{}, err
	}
	return &addresses, nil
}

// DeleteAddress removes an address assignment from the DB (should never use this unless correcting an accidental addition)
func (aa *AddressAssignment) DeleteAddress(db *gorm.DB, aaid uint64) error {

	err := db.Debug().Model(&AddressAssignment{}).Where("id = ?", aa.ID).Updates(AddressAssignment{Status: Deleted, UpdatedAt: time.Now()}).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("Address not found")
		}
		return err
	}

	if aa.Status == Permanent {
		priorAddress := AddressAssignment{}
		err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND status = ? AND end_date = ?", aa.UserID, Permanent, aa.StartDate.AddDate(0, 0, -1)).Find(&priorAddress).Error
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return errors.New("Could not find previous address")
			}
			return err
		}
		newEndDate := aa.EndDate
		if newEndDate.Valid {
			err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND id = ?", aa.UserID, priorAddress.ID).Updates(AddressAssignment{EndDate: newEndDate, UpdatedAt: time.Now()}).Error
			if err != nil {
				if gorm.IsRecordNotFoundError(err) {
					return errors.New("Address not found")
				}
				return db.Error
			}
		} else {
			err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND id = ?", aa.UserID, priorAddress.ID).Updates(map[string]interface{}{"end_date": nil, "updated_at": time.Now()}).Error
			if err != nil {
				if gorm.IsRecordNotFoundError(err) {
					return errors.New("Address not found")
				}
				return db.Error
			}
		}

	}

	return nil
}

// DeleteAddressAssignment removes an address assignment from the DB (should never use this unless correcting an accidental addition)
func (aa *AddressAssignment) DeleteAddressAssignment(db *gorm.DB, aaid uint64) (int64, error) {

	db = db.Debug().Model(&AddressAssignment{}).Where("id = ?", aaid).Take(&AddressAssignment{}).Delete(&AddressAssignment{})

	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Address not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
