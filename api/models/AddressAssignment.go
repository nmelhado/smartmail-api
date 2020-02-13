package models

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
)

// Status is an enum type used for AddressAssignment (valies cannot be added wothot altering the DB first!)
// refer to link for `Status` field: https://github.com/jinzhu/gorm/issues/1978
type Status string

const (
	longTerm             Status = "long_term"
	temporary            Status = "temporary"
	packageOnlyLongTerm  Status = "package_only_long_term"
	packageOnlyTemporary Status = "package_only_temporary"
	mailOnlyLongTerm     Status = "mail_only_long_term"
	mailOnlyTemporary    Status = "mail_only_temporary"
	expired              Status = "expired"
	deleted              Status = "deleted"
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
	longTerm,
	temporary,
	packageOnlyLongTerm,
	packageOnlyTemporary,
}

var validMailStatus []Status = []Status{
	longTerm,
	temporary,
	mailOnlyLongTerm,
	mailOnlyTemporary,
}

var temporaryStatus []Status = []Status{
	temporary,
	packageOnlyTemporary,
	mailOnlyTemporary,
}

var longTermStatus []Status = []Status{
	longTerm,
	mailOnlyLongTerm,
	packageOnlyLongTerm,
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
	if !aa.StartDate.IsZero() {
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
		err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND status IN ('temporary', ?) AND ((start_date <= ? AND start_date >= ?) OR (end_date >= ? AND end_date <= ?))", aa.UserID, aa.Status, aa.EndDate, aa.StartDate, aa.StartDate, aa.EndDate).Limit(100).Find(&conflictingAddresses).Error
		if err != nil {
			return &AddressAssignment{}, err
		}
		if len(conflictingAddresses) > 0 {
			return &AddressAssignment{}, errors.New("conflict with another temporary address - please make sure that the dates for temporary addresses don't overlap")
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
		if contains(longTermStatus, aa.Status) {
			err = db.Debug().Model(&AddressAssignment{}).Where("status IN (?) AND id <> ? AND user_id = ?", longTermStatus, aa.ID, aa.UserID).Updates(AddressAssignment{EndDate: null.TimeFrom(aa.StartDate), UpdatedAt: time.Now()}).Error
		}
	}
	return aa, nil
}

// UpdateAddressAssignment is used to update status as well as start and end dates.
func (aa *AddressAssignment) UpdateAddressAssignment(db *gorm.DB) (*AddressAssignment, error) {

	var err error
	err = db.Debug().Model(&AddressAssignment{}).Where("id = ?", aa.ID).Updates(AddressAssignment{Status: aa.Status, StartDate: aa.StartDate, EndDate: aa.EndDate, UpdatedAt: time.Now()}).Error
	if err != nil {
		return &AddressAssignment{}, err
	}
	return aa, nil
}

// FindMailingAddressWithCosmo allows a mailcarrier to retrieve the correct address to send mail to a user by inputing a User (retieved through CosmoID) and an estimated date of delivery
func (aa *AddressAssignment) FindMailingAddressWithCosmo(db *gorm.DB, user User, targetDate time.Time) (*AddressAssignment, error) {
	var err error
	address := AddressAssignment{}

	err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND status IN (?, ?) AND start_date < ? AND (end_date IS NULL OR end_date > ?)", user.ID, mailOnlyTemporary, temporary, targetDate, targetDate).Find(&address).Error
	if err != nil {
		return &AddressAssignment{}, err
	}

	if address.ID == 0 {

		err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND status IN (?) AND start_date < ? AND (end_date IS NULL OR end_date > ?)", user.ID, validMailStatus, targetDate, targetDate).Find(&address).Error
		if err != nil {
			return &AddressAssignment{}, err
		}

	}

	if address.ID > 0 {
		err := db.Debug().Model(&Address{}).Where("id = ?", address.AddressID).Take(&address.Address).Error
		if err != nil {
			return &AddressAssignment{}, err
		}
		address.User = user
	}
	return &address, nil
}

// FindPackageAddressWithCosmo allows a mailcarrier to retrieve the correct address to send packages to a user by inputing a User (retieved through CosmoID) and an estimated date of delivery
func (aa *AddressAssignment) FindPackageAddressWithCosmo(db *gorm.DB, user User, targetDate time.Time) (*AddressAssignment, error) {
	var err error
	address := AddressAssignment{}

	err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND status IN (?, ?) AND start_date < ? AND (end_date IS NULL OR end_date > ?)", user.ID, packageOnlyTemporary, temporary, targetDate, targetDate).Find(&address).Error
	if err != nil {
		return &AddressAssignment{}, err
	}

	if address.ID == 0 {

		err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND status IN (?) AND start_date < ? AND (end_date IS NULL OR end_date > ?)", user.ID, validPackageStatus, targetDate, targetDate).Find(&address).Error
		if err != nil {
			return &AddressAssignment{}, err
		}

	}

	if address.ID > 0 {
		err := db.Debug().Model(&Address{}).Where("id = ?", address.AddressID).Take(&address.Address).Error
		if err != nil {
			return &AddressAssignment{}, err
		}
		address.User = user
	}
	return &address, nil
}

// FindAllAddressesForUser retieves the last 100 addresses a user has linked to their account. Used in UI to provide address history
func (aa *AddressAssignment) FindAllAddressesForUser(db *gorm.DB, uid uint64) (*[]AddressAssignment, error) {
	var err error
	addresses := []AddressAssignment{}
	err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND status <> ?", uid, deleted).Limit(100).Find(&addresses).Error
	if err != nil {
		return &[]AddressAssignment{}, err
	}
	if len(addresses) > 0 {
		for i := range addresses {
			err := db.Debug().Model(&User{}).Where("id = ?", addresses[i].UserID).Take(&addresses[i].User).Error
			if err != nil {
				return &[]AddressAssignment{}, err
			}
			err = db.Debug().Model(&Address{}).Where("id = ?", addresses[i].AddressID).Take(&addresses[i].Address).Error
			if err != nil {
				return &[]AddressAssignment{}, err
			}
		}
	}
	return &addresses, nil
}
