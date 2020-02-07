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
	// not how enum works, view bookmarked page
	Status   enum    `gorm:"default:false;" json:"temporary"`
	StartDate time.Time `gorm:"default:CURRENT_TIMESTAMP;not null;" json:"start_date"`
	EndDate time.Time `gorm:"default:null" json:"start_date"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

var invalidStatus []string = []string{
	"inactive",
	"deleted"
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

// need to create validate function

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
	// db = db.Debug().Model(&Post{}).Where("id = ?", pid).Take(&Post{}).UpdateColumns(
	// 	map[string]interface{}{
	// 		"title":      p.Title,
	// 		"content":    p.Content,
	// 		"updated_at": time.Now(),
	// 	},
	// )
	// err = db.Debug().Model(&Post{}).Where("id = ?", pid).Take(&p).Error
	// if err != nil {
	// 	return &Post{}, err
	// }
	// if p.ID != 0 {
	// 	err = db.Debug().Model(&User{}).Where("id = ?", p.AuthorID).Take(&p.Author).Error
	// 	if err != nil {
	// 		return &Post{}, err
	// 	}
	// }
	err = db.Debug().Model(&AddressAssignment{}).Where("id = ?", aa.ID).Updates(AddressAssignment{Status: aa.Status, StartDate: aa.StartDate, EndDate: aa.EndDate, UpdatedAt: time.Now()}).Error
	if err != nil {
		return &AddressAssignment{}, err
	}
	return aa, nil
}

func (aa *AddressAssignment) FindMailingAddressWithCosmo(db *gorm.DB, cosmoID string, targetDate time.date) (*Address, error) {
	var err error
	var user User{}
	address := AddressAssignment{}
	
	err = db.Debug().Model(&User{}).Where("ix_cosmo_id = ?", cosmoID).Take(&user).Error
	if err != nil {
		return &Address{}, err
	}
	
	
	err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ? AND staus NOT IN ? AND ? BETWEEN star_date AND end_date", user.id, invalidStatus, targetDate).Find(&address).Error
	if err != nil {
		return &Address{}, err
	}
	if address.ID > 0 {
		err := db.Debug().Model(&Address{}).Where("id = ?", address.AddressID).Take(&address.Address).Error
		if err != nil {
			return &Address{}, err
		}
		&address.User = user
	}
	return &address, nil
}

func (aa *AddressAssignment) FindAllAddressesForUser(db *gorm.DB, uid uint64) (*[]AddressAssignment, error) {
	var err error
	addresses := []AddressAssignment{}
	err = db.Debug().Model(&AddressAssignment{}).Where("user_id = ?", uid).Limit(100).Find(&addresses).Error
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
