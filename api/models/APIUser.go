package models

import (
	"database/sql/driver"
	"errors"
	"html"
	"log"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type Permission string

/*
postgres command to create enum:
CREATE TYPE api_permission AS ENUM (
	'none',
	'mail_carrier',
	'retailer',
	'admin',
	'engineer');
*/

const (
	// NoPermission is the default permission type, it gives no access
	NoPermission Permission = "none"
	// MailCarrierPermission is the mail carrier permission type, it allows the user to get zip code and address information
	MailCarrierPermission Permission = "mail_carrier"
	// RetailerPermission is the retailer permission type, it allows the user to get zip code information
	RetailerPermission Permission = "retailer"
	// AdminPermission is the admin permission type, it allows the user all information
	AdminPermission Permission = "admin"
	// EngineerPermission is the admin permission type, it allows the user all information (may be limited in the future)
	EngineerPermission Permission = "engineer"
)

func (p *Permission) Scan(value interface{}) error {
	*p = Permission(value.([]byte))
	return nil
}

// Value returns the value of the authority enum
func (p Permission) Value() (driver.Value, error) {
	return string(p), nil
}

// APIUser is the DB and json structure for an API user
type APIUser struct {
	ID              uuid.UUID     `gorm:"type:uuid;primary_key" json:"id"`
	Username        string        `gorm:"size:100;not null;unique" json:"username"`
	Email           string        `gorm:"size:100;not null;unique" json:"email"`
	Name            string        `gorm:"size:30;not null;" json:"name"`
	Phone           string        `gorm:"size:30;not null;" json:"phone"`
	SmartmailUser   User          `json:"smartmail_user"`
	SmartmailUserID uuid.NullUUID `gorm:"type:uuid;" sql:"type:uuid REFERENCES users(id)" json:"smartmail_user_id"`
	Permission      Permission    `gorm:"default:'none'" sql:"type:api_permission" json:"permission"`
	Password        string        `gorm:"size:100;not null;" json:"password"`
	CreatedAt       time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// FullAddressPermissions returns true if the provided permission level is adequate to access address full information
func FullAddressPermissions(permission Permission) bool {
	switch permission {
	case
		AdminPermission,
		EngineerPermission,
		MailCarrierPermission:
		return true
	}
	return false
}

// ZipPermissions returns true if the provided permission level is adequate to access zip code information
func ZipPermissions(permission Permission) bool {
	switch permission {
	case
		AdminPermission,
		EngineerPermission,
		MailCarrierPermission,
		RetailerPermission:
		return true
	}
	return false
}

// RetailPermissions returns true if the provided permission level is adequate to access zip code information
func RetailPermissions(permission Permission) bool {
	switch permission {
	case
		AdminPermission,
		EngineerPermission,
		RetailerPermission:
		return true
	}
	return false
}

// BeforeCreate will set a UUID rather than numeric ID.
func (au *APIUser) BeforeCreate(scope *gorm.Scope) error {
	uuid := uuid.NewV4()
	return scope.SetColumn("ID", uuid)
}

// Hash and VerifyPassword are located in the APIUser model

// BeforeSave converts a string password into a hashed oassword before uploading to the DB
func (au *APIUser) BeforeSave() error {
	hashedPassword, err := Hash(au.Password)
	if err != nil {
		return err
	}
	au.Password = string(hashedPassword)
	return nil
}

// Prepare sanitizes an API user object before other operations are performed
func (au *APIUser) Prepare() {
	au.ID = uuid.UUID{}
	au.Username = html.EscapeString(strings.ToUpper(strings.TrimSpace(au.Username)))
	au.Name = html.EscapeString(strings.TrimSpace(au.Name))
	au.Email = html.EscapeString(strings.ToLower(strings.TrimSpace(au.Email)))
	au.Phone = html.EscapeString(strings.TrimSpace(au.Phone))
	au.Permission = NoPermission // Set the permission to none and then update it to the appropriate permission level
	au.CreatedAt = time.Now()
	au.UpdatedAt = time.Now()
}

// Validate ensures proper inputs
func (au *APIUser) Validate(action string) error {
	switch strings.ToLower(action) {
	case "update":
		if au.Username == "" {
			return errors.New("Username required")
		}
		if au.Name == "" {
			return errors.New("Name required")
		}
		if au.Password == "" {
			return errors.New("Password required")
		}
		if au.Phone == "" {
			return errors.New("Phone required")
		}
		if au.Email == "" {
			return errors.New("Email required")
		}
		if err := checkmail.ValidateFormat(au.Email); err != nil {
			return errors.New("Invalid email")
		}
		return nil
	case "login":
		if au.Password == "" {
			return errors.New("Password required")
		}
		if au.Username == "" {
			return errors.New("Username required")
		}
		return nil

	default:
		if au.Username == "" {
			return errors.New("Username required")
		}
		if au.Name == "" {
			return errors.New("Name required")
		}
		if au.Password == "" {
			return errors.New("Password required")
		}
		if au.Phone == "" {
			return errors.New("Phone required")
		}
		if au.Email == "" {
			return errors.New("Email required")
		}
		if err := checkmail.ValidateFormat(au.Email); err != nil {
			return errors.New("Invalid email")
		}
		return nil
	}
}

// SaveAPIUser saves an API user to the DB
func (au *APIUser) SaveAPIUser(db *gorm.DB) (*APIUser, error) {

	var err error
	err = db.Debug().Create(&au).Error
	if err != nil {
		return &APIUser{}, err
	}
	return au, nil
}

// FindAllAPIUsers retrieves 100 API users from the DB
func (au *APIUser) FindAllAPIUsers(db *gorm.DB) (*[]APIUser, error) {
	var err error
	mailCarriers := []APIUser{}
	err = db.Debug().Model(&APIUser{}).Limit(100).Find(&mailCarriers).Error
	if err != nil {
		return &[]APIUser{}, err
	}
	return &mailCarriers, err
}

// FindAPIUserByID retrieves the data for an API user by using their ID
func (au *APIUser) FindAPIUserByID(db *gorm.DB, uid uuid.UUID) (*APIUser, error) {
	var err error
	err = db.Debug().Set("gorm:auto_preload", true).Model(APIUser{}).Where("id = ?", uid).Take(&au).Error
	if err != nil {
		return &APIUser{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &APIUser{}, errors.New("User Not Found")
	}
	return au, err
}

// UpdateAPIUser updates the values of an API user
func (au *APIUser) UpdateAPIUser(db *gorm.DB, uid uuid.UUID) (*APIUser, error) {

	// To hash the password
	err := au.BeforeSave()
	if err != nil {
		log.Fatal(err)
	}
	db = db.Debug().Model(&APIUser{}).Where("id = ?", uid).Take(&APIUser{}).UpdateColumns(
		map[string]interface{}{
			"password":  au.Password,
			"name":      au.Name,
			"username":  au.Username,
			"phone":     au.Phone,
			"email":     au.Email,
			"update_at": time.Now(),
		},
	)
	if db.Error != nil {
		return &APIUser{}, db.Error
	}
	// This is to display the updated API user
	err = db.Debug().Model(&APIUser{}).Where("id = ?", uid).Take(&au).Error
	if err != nil {
		return &APIUser{}, err
	}
	return au, nil
}

// DeleteAPIUser removes an API user from the DB
func (au *APIUser) DeleteAPIUser(db *gorm.DB, uid uuid.UUID) (int64, error) {

	db = db.Debug().Model(&APIUser{}).Where("id = ?", uid).Take(&APIUser{}).Delete(&APIUser{})

	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
