package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"gopkg.in/guregu/null.v3"
)

// Address is the sql and and json model for an address
type Address struct {
	ID                   uint64      `gorm:"primary_key;auto_increment" json:"id"`
	Nickname             null.String `gorm:"size:255;" json:"nickname"`
	LineOne              string      `gorm:"size:255;not null;" json:"line_one"`
	LineTwo              null.String `gorm:"size:255;" json:"line_two"`
	BusinessName         null.String `gorm:"size:255;" json:"business_name"`
	AttentionTo          null.String `gorm:"size:255;" json:"attention_to"`
	City                 string      `gorm:"size:255;not null;" json:"city"`
	State                string      `gorm:"size:255;not null;" json:"state"`
	ZipCode              string      `gorm:"size:255;not null;" json:"zip_code"`
	Country              string      `gorm:"size:255;not null;" json:"country"`
	Phone                null.String `gorm:"size:255;" json:"phone"`
	Latitude             float64     `json:"latitude"`
	Longitude            float64     `json:"longitude"`
	DeliveryInstructions null.String `gorm:"size:255;" json:"delivery_instructions"`
	CreatedAt            time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// Prepare escapes html, trims strings, and sets created and updated times for the various Address fields
func (a *Address) Prepare() {
	a.ID = 0
	a.Nickname.String = html.UnescapeString(strings.TrimSpace(a.Nickname.String))
	a.LineOne = html.UnescapeString(strings.TrimSpace(a.LineOne))
	a.LineTwo.String = html.UnescapeString(strings.TrimSpace(a.LineTwo.String))
	a.BusinessName.String = html.UnescapeString(strings.TrimSpace(a.BusinessName.String))
	a.AttentionTo.String = html.UnescapeString(strings.TrimSpace(a.AttentionTo.String))
	a.City = html.UnescapeString(strings.TrimSpace(a.City))
	a.State = html.UnescapeString(strings.TrimSpace(a.State))
	a.ZipCode = html.UnescapeString(strings.TrimSpace(a.ZipCode))
	a.Country = html.UnescapeString(strings.TrimSpace(a.Country))
	a.Phone.String = html.UnescapeString(strings.TrimSpace(a.Phone.String))
	a.DeliveryInstructions.String = html.UnescapeString(strings.TrimSpace(a.DeliveryInstructions.String))
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
}

// Validate ensures that proper input were received
func (a *Address) Validate() error {

	if a.LineOne == "" {
		return errors.New("Address Line 1 required")
	}
	if a.City == "" {
		return errors.New("City required")
	}
	if a.State == "" {
		return errors.New("State required")
	}
	if a.ZipCode == "" {
		return errors.New("Zip Code required")
	}
	if a.Country == "" {
		return errors.New("Country required")
	}
	return nil
}

// SaveAddress saves an address. This is typically followed immediately by a SaveAddressAssignment call.
func (a *Address) SaveAddress(db *gorm.DB) (*Address, error) {
	var err error
	err = db.Debug().Model(&Address{}).Create(&a).Error
	if err != nil {
		return &Address{}, err
	}
	return a, nil
}

// Below still needs revision

// Below function may be unnecessary (comment it out, when back at a computer, for now)

// func (p *Address) FindAllPosts(db *gorm.DB) (*[]Post, error) {
// 	var err error
// 	posts := []Post{}
// 	err = db.Debug().Model(&Post{}).Limit(100).Find(&posts).Error
// 	if err != nil {
// 		return &[]Post{}, err
// 	}
// 	if len(posts) > 0 {
// 		for i, _ := range posts {
// 			err := db.Debug().Model(&User{}).Where("id = ?", posts[i].AuthorID).Take(&posts[i].Author).Error
// 			if err != nil {
// 				return &[]Post{}, err
// 			}
// 		}
// 	}
// 	return &posts, nil
// }

// FindAddressByID find an address using its ID value
func (a *Address) FindAddressByID(db *gorm.DB, aid uint64) (*Address, error) {
	var err error
	err = db.Debug().Model(&Address{}).Where("id = ?", aid).Take(&a).Error
	if err != nil {
		return &Address{}, err
	}
	return a, nil
}

// Update updates the values of an address (this is for correcting an address, NOT updating a user's current address)
func (a *Address) Update(db *gorm.DB, aid uint64) error {
	var err error

	err = db.Debug().Model(&Address{}).Where("id = ?", aid).Updates(
		Address{
			Nickname:             a.Nickname,
			LineOne:              a.LineOne,
			LineTwo:              a.LineTwo,
			BusinessName:         a.BusinessName,
			AttentionTo:          a.AttentionTo,
			City:                 a.City,
			State:                a.State,
			ZipCode:              a.ZipCode,
			Country:              a.Country,
			Phone:                a.Phone,
			Latitude:             a.Latitude,
			Longitude:            a.Longitude,
			DeliveryInstructions: a.DeliveryInstructions,
			UpdatedAt:            time.Now()}).Error
	if err != nil {
		return err
	}
	return nil
}

// DeleteAddress removes an address from the DB (should never use this unless correcting an accidental addition)
func (a *Address) DeleteAddress(db *gorm.DB, aid uint64) (int64, error) {

	db = db.Debug().Model(&Address{}).Where("id = ?", aid).Delete(&Address{})

	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Address not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
