package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type AddressAssignment struct {
	ID        uint64    `gorm:"primary_key;auto_increment" json:"id"`
	User    User      `json:"user"`
	UserID  uint32    `sql:"type:int REFERENCES users(id)" json:"user_id"`
	Address    Address      `json:"address"`
	AddressID  uint32    `sql:"type:int REFERENCES addresses(id)" json:"address_id"`
	Status   enum    `gorm:"default:false;" json:"temporary"`
	StartDate time.Time `gorm:"default:CURRENT_TIMESTAMP;not null;" json:"start_date"`
	EndDate time.Time `gorm:"default:null" json:"start_date"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (aa *AddressAssignment) Prepare() {
	aa.ID = 0
	aa.User = User{}
	aa.Address = Address{}

	aa.Status = html.EscapeString(strings.TrimSpace(aa.Status))
	
	aa.StartDate = time.Now()
	aa.EndDate = time.Now()
	aa.CreatedAt = time.Now()
	aa.UpdatedAt = time.Now()
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
