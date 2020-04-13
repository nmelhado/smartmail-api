package models

import (
	"errors"
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

// GetContacts retrieves all of a user's contacts
func GetContacts(db *gorm.DB, userID uuid.UUID) (contacts []Contact, err error) {
	err = db.Debug().Model(&Contact{}).Where("user_id = ?", userID).Limit(100).Find(&contacts).Error
	if err != nil {
		return []Contact{}, err
	}
	if len(contacts) > 0 {
		for i := range contacts {
			err := db.Debug().Model(&User{}).Where("id = ?", contacts[i].ContactID).Take(&contacts[i].Contact).Error
			if err != nil {
				return []Contact{}, err
			}
		}
	}
	return contacts, nil
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
