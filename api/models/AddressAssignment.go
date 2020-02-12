package models

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
)

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

func (s *Status) Scan(value interface{}) error {
	*s = Status(value.([]byte))
	return nil
}

func (s Status) Value() (driver.Value, error) {
	return string(s), nil
}

// refer to link for `Status` field: https://github.com/jinzhu/gorm/issues/1978
type AddressAssignment struct {
	ID        uint64    `gorm:"primary_key;auto_increment" json:"id"`
	User      User      `json:"user"`
	UserID    uuid.UUID `gorm:"type:uuid" sql:"type:uuid REFERENCES users(id)" json:"user_id"`
	Address   Address   `json:"address"`
	AddressID uint64    `sql:"type:int REFERENCES addresses(id)" json:"address_id"`
	Status    Status    `sql:"type:status" json:"status"`
	StartDate null.Time `gorm:"default:CURRENT_TIMESTAMP;not null;" json:"start_date"`
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

func (aa *AddressAssignment) Prepare() {
	aa.ID = 0
	aa.User = User{}
	aa.Address = Address{}
	aa.CreatedAt = time.Now()
	aa.UpdatedAt = time.Now()
}

func (aa *AddressAssignment) Validate() error {
	if status, err := aa.Status.Value(); status == "" || err != nil {
		return errors.New("Status required")
	}
	if !aa.StartDate.Valid {
		return errors.New("Start date required")
	}
	if contains(temporaryStatus, aa.Status) {
		if aa.EndDate.Valid {
			return errors.New("End date required for temporary address")
		}
	}
	return nil
}

func (aa *AddressAssignment) SaveAddressAssignment(db *gorm.DB) (*AddressAssignment, error) {
	var err error
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
		err = db.Debug().Model(&AddressAssignment{}).Where("status IN (?) AND id <> ? AND user_id = ?", longTermStatus, aa.ID, aa.UserID).Updates(AddressAssignment{Status: expired, EndDate: aa.StartDate, UpdatedAt: time.Now()}).Error
	}
	return aa, nil
}

func (aa *AddressAssignment) UpdateAddressAssignment(db *gorm.DB) (*AddressAssignment, error) {

	var err error
	err = db.Debug().Model(&AddressAssignment{}).Where("id = ?", aa.ID).Updates(AddressAssignment{Status: aa.Status, StartDate: aa.StartDate, EndDate: aa.EndDate, UpdatedAt: time.Now()}).Error
	if err != nil {
		return &AddressAssignment{}, err
	}
	return aa, nil
}

func (aa *AddressAssignment) FindMailingAddressWithCosmo(db *gorm.DB, user User, targetDate time.Time) (*AddressAssignment, error) {
	var err error
	address := AddressAssignment{}

	err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND status IN (?) AND start_date < ? AND (end_date IS NULL OR end_date > ?)", user.ID, validMailStatus, targetDate, targetDate).Find(&address).Error
	if err != nil {
		return &AddressAssignment{}, err
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

func (aa *AddressAssignment) FindPackageAddressWithCosmo(db *gorm.DB, user User, targetDate time.Time) (*AddressAssignment, error) {
	var err error
	address := AddressAssignment{}

	err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND status IN (?) AND start_date < ? AND (end_date IS NULL OR end_date > ?)", user.ID, validPackageStatus, targetDate, targetDate).Find(&address).Error
	if err != nil {
		return &AddressAssignment{}, err
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

func (aa *AddressAssignment) FindAllAddressesForUser(db *gorm.DB, uid uint64) (*[]AddressAssignment, error) {
	var err error
	addresses := []AddressAssignment{}
	err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND status <> ?", uid, deleted).Limit(100).Find(&addresses).Error
	if err != nil {
		return &[]AddressAssignment{}, err
	}
	if len(addresses) > 0 {
		for i, _ := range addresses {
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