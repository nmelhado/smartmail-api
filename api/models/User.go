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
	"golang.org/x/crypto/bcrypt"
)

type authority string

const (
	UserAuth     authority = "user"
	MailerAuth   authority = "mailer"
	AdminAuth    authority = "admin"
	EngineerAuth authority = "engineer"
	RetailerAuth authority = "retailer"
)

func (a *authority) Scan(value interface{}) error {
	*a = authority(value.([]byte))
	return nil
}

// Value returns the value of the authority enum
func (a authority) Value() (driver.Value, error) {
	return string(a), nil
}

// User is the DB and json structure for a user
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	CosmoID   string    `gorm:"size:8;not null;unique;unique_index:ix_cosmo_id" json:"cosmo_id"`
	Email     string    `gorm:"size:100;not null;unique" json:"email"`
	FirstName string    `gorm:"size:30;not null;" json:"first_name"`
	LastName  string    `gorm:"size:30;not null;" json:"last_name"`
	Phone     string    `gorm:"size:30;not null;unique" json:"phone"`
	Authority authority `sql:"type:authority" json:"authority"`
	Password  string    `gorm:"size:100;not null;" json:"password"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (u *User) BeforeCreate(scope *gorm.Scope) error {
	uuid := uuid.NewV4()
	return scope.SetColumn("ID", uuid)
}

// Hash creates a hass of the user's provided oassword
func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// VerifyPassword compares the provided password to the hashed password stored in the DB
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// BeforeSave converts a string password into a hashed oassword before uploading to the DB
func (u *User) BeforeSave() error {
	hashedPassword, err := Hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// Prepare sanitizes a user object before other operations are performed
func (u *User) Prepare() {
	u.ID = uuid.UUID{}
	u.CosmoID = html.EscapeString(strings.ToUpper(strings.TrimSpace(u.CosmoID)))
	u.FirstName = html.EscapeString(strings.TrimSpace(u.FirstName))
	u.LastName = html.EscapeString(strings.TrimSpace(u.LastName))
	u.Email = html.EscapeString(strings.ToLower(strings.TrimSpace(u.Email)))
	u.Phone = html.EscapeString(strings.TrimSpace(u.Phone))
	u.Authority = UserAuth
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
}

// Validate ensures proper inputs
func (u *User) Validate(action string) error {
	switch strings.ToLower(action) {
	case "update":
		if u.FirstName == "" {
			return errors.New("First name required")
		}
		if u.LastName == "" {
			return errors.New("Last name required")
		}
		if u.Password == "" {
			return errors.New("Password required")
		}
		if u.Phone == "" {
			return errors.New("Phone required")
		}
		if u.Email == "" {
			return errors.New("Email required")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("Invalid email")
		}
		return nil
	case "login":
		if u.Password == "" {
			return errors.New("Password required")
		}
		if u.Email == "" {
			return errors.New("Email required")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("Invalid Email")
		}
		return nil

	default:
		if u.FirstName == "" {
			return errors.New("First name required")
		}
		if u.LastName == "" {
			return errors.New("Last name required")
		}
		if u.Password == "" {
			return errors.New("Password required")
		}
		if u.Phone == "" {
			return errors.New("Phone required")
		}
		if u.Email == "" {
			return errors.New("Email required")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("Invalid email")
		}
		return nil
	}
}

// Saves a user to the DB. Almost always done in conjunction with saving a user's first address and address assignment
func (u *User) SaveUser(db *gorm.DB) (*User, error) {

	var err error
	err = db.Debug().Create(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

// FindAllUsers retrieves 100 users from the DB
func (u *User) FindAllUsers(db *gorm.DB) (*[]User, error) {
	var err error
	users := []User{}
	err = db.Debug().Model(&User{}).Limit(100).Find(&users).Error
	if err != nil {
		return &[]User{}, err
	}
	return &users, err
}

func (u *User) FindUserByID(db *gorm.DB, uid uuid.UUID) (*User, error) {
	var err error
	err = db.Debug().Model(User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &User{}, errors.New("User Not Found")
	}
	return u, err
}

func (u *User) UpdateAUser(db *gorm.DB, uid uuid.UUID) (*User, error) {

	// To hash the password
	err := u.BeforeSave()
	if err != nil {
		log.Fatal(err)
	}
	db = db.Debug().Model(&User{}).Where("id = ?", uid).Take(&User{}).UpdateColumns(
		map[string]interface{}{
			"password":   u.Password,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"phone":      u.Phone,
			"email":      u.Email,
			"update_at":  time.Now(),
		},
	)
	if db.Error != nil {
		return &User{}, db.Error
	}
	// This is the display the updated user
	err = db.Debug().Model(&User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

func (u *User) DeleteUser(db *gorm.DB, uid uuid.UUID) (int64, error) {

	db = db.Debug().Model(&User{}).Where("id = ?", uid).Take(&User{}).Delete(&User{})

	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
