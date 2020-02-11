package models

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
)

type status string

const (
	longTerm             status = "long_term"
	temporary            status = "temporary"
	packageOnlyLongTerm  status = "package_only_long_term"
	packageOnlyTemporary status = "package_only_temporary"
	mailOnlyLongTerm     status = "mail_only_long_term"
	mailOnlyTemporary    status = "mail_only_temporary"
	expired              status = "expired"
	deleted              status = "deleted"
)

func (s *status) Scan(value interface{}) error {
	*s = status(value.([]byte))
	return nil
}

func (s status) Value() (driver.Value, error) {
	return string(s), nil
}

// refer to link for `status` field: https://github.com/jinzhu/gorm/issues/1978
type AddressAssignment struct {
	ID        uint64    `gorm:"primary_key;auto_increment" json:"id"`
	User      User      `json:"user"`
	UserID    uuid.UUID `gorm:"type:uuid" sql:"type:uuid REFERENCES users(id)" json:"user_id"`
	Address   Address   `json:"address"`
	AddressID uint64    `sql:"type:int REFERENCES addresses(id)" json:"address_id"`
	Status    status    `sql:"type:status" json:"status"`
	StartDate null.Time `gorm:"default:CURRENT_TIMESTAMP;not null;" json:"start_date"`
	EndDate   null.Time `gorm:"default:null" json:"end_date"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

var validPackageStatus []status = []status{
	longTerm,
	temporary,
	packageOnlyLongTerm,
	packageOnlyTemporary,
}

var validMailStatus []status = []status{
	longTerm,
	temporary,
	mailOnlyLongTerm,
	mailOnlyTemporary,
}

var temporaryStatus []status = []status{
	temporary,
	packageOnlyTemporary,
	mailOnlyTemporary,
}

func contains(arr []status, status status) bool {
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

	err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND staus IN ? AND ? BETWEEN star_date AND end_date", user.ID, validMailStatus, targetDate).Find(&address).Error
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

	err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND staus IN ? AND ? BETWEEN star_date AND end_date", user.ID, validPackageStatus, targetDate).Find(&address).Error
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
