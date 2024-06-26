package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// Contact is the DB table structure and json input structure for a contact in a user's address book. It is a many to many relationship table.
type Contact struct {
	ID        uint64    `gorm:"primary_key;auto_increment" json:"id"`
	User      User      `json:"user"`
	UserID    uuid.UUID `gorm:"type:uuid" sql:"type:uuid REFERENCES users(id)" json:"user_id"`
	Contact   User      `json:"contact"`
	ContactID uuid.UUID `gorm:"type:uuid" sql:"type:uuid REFERENCES users(id)" json:"contact_id"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// Prepare formats the Contact object
func (c *Contact) Prepare() {
	c.ID = 0
	c.User = User{}
	c.Contact = User{}
	c.CreatedAt = time.Now()
}

// SaveContacts is used to save a contact to both the recipient and sender
func (c *Contact) SaveContacts(db *gorm.DB, userID uuid.UUID, contactID uuid.UUID) (err error) {
	contact := Contact{}
	err = db.Debug().Model(&Contact{}).Where("user_id = ? AND contact_id = ?", userID, contactID).Attrs(Contact{UserID: userID, ContactID: contactID, CreatedAt: time.Now()}).FirstOrCreate(&contact).Error
	if err != nil {
		return err
	}

	secondContact := Contact{}
	err = db.Debug().Model(&Contact{}).Where("user_id = ? AND contact_id = ?", contactID, userID).Attrs(Contact{UserID: contactID, ContactID: userID, CreatedAt: time.Now()}).FirstOrCreate(&secondContact).Error
	if err != nil {
		return err
	}
	return nil
}

// SaveContact is used to save a contact to an avvount
func (c *Contact) SaveContact(db *gorm.DB, userID uuid.UUID, contactID uuid.UUID) (contact Contact, err error) {
	err = db.Debug().Model(&Contact{}).Where("user_id = ? AND contact_id = ?", userID, contactID).Attrs(Contact{UserID: userID, ContactID: contactID, CreatedAt: time.Now()}).FirstOrCreate(&contact).Error
	if err != nil {
		return Contact{}, err
	}
	err = db.Debug().Set("gorm:auto_preload", true).Model(&Contact{}).Take(&contact).Error

	return
}

// GetContacts retrieves all of a user's contacts
func GetContacts(db *gorm.DB, userID uuid.UUID, limit int64, offset int64, sortPref string) (count int64, contacts []Contact, err error) {
	if sortPref == "recent" {
		err = db.Debug().Model(&Contact{}).Order("created_at desc").Where("user_id = ?", userID).Count(&count).Limit(limit).Offset(offset).Preload("Contact").Find(&contacts).Error
		if err != nil {
			return 0, []Contact{}, err
		}
	} else {
		err = db.Debug().Model(&Contact{}).Joins("left join users on users.id = contacts.contact_id").Order("users.first_name").Order("users.last_name").Where("user_id = ?", userID).Count(&count).Limit(limit).Offset(offset).Preload("Contact").Find(&contacts).Error
		if err != nil {
			return 0, []Contact{}, err
		}
	}
	return
}

// SearchContacts retrieves all of a user's contacts that meet a user's search parameter
func SearchContacts(db *gorm.DB, userID uuid.UUID, limit int64, offset int64, search string) (count int64, contacts []Contact, err error) {
	searchTerms := []string{}
	rawSearchTerms := strings.Fields(search)
	for _, singleSearchTerm := range rawSearchTerms {
		searchTerms = append(searchTerms, strings.ToLower(singleSearchTerm))
	}
	if len(searchTerms) == 1 {
		likeTerm := fmt.Sprintf("%%%s%%", searchTerms[0])
		err = db.Debug().Model(&Contact{}).Joins("left join users on users.id = contacts.contact_id").Order("users.first_name").Order("users.last_name").Where("user_id = ? AND (LOWER(users.first_name) LIKE ? OR LOWER(users.last_name) LIKE ? OR LOWER(users.smart_id) = ?)", userID, likeTerm, likeTerm, searchTerms[0]).Count(&count).Limit(limit).Offset(offset).Preload("Contact").Find(&contacts).Error
	} else {
		err = db.Debug().Model(&Contact{}).Joins("left join users on users.id = contacts.contact_id").Order("users.first_name").Order("users.last_name").Where("user_id = ? AND (LOWER(users.first_name) IN (?) OR LOWER(users.last_name) IN (?) OR LOWER(users.smart_id) IN (?))", userID, searchTerms, searchTerms, searchTerms).Count(&count).Limit(limit).Offset(offset).Preload("Contact").Find(&contacts).Error
	}
	if err != nil {
		return 0, []Contact{}, err
	}
	return
}

// DeleteContact removes a contact from the DB
func (c *Contact) DeleteContact(db *gorm.DB, cid uint64) (int64, error) {

	db = db.Debug().Model(&Contact{}).Where("id = ?", cid).Take(&Contact{}).Delete(&Contact{})

	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Contact not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
