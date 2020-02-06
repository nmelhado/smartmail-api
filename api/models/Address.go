package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type Address struct {
	ID        uint64    `gorm:"primary_key;auto_increment" json:"id"`
	Nickname   string    `gorm:"size:255;" json:"nickname"`
	LineOne     string    `gorm:"size:255;not null;" json:"line_one"`
	LineTwo   string    `gorm:"size:255;" json:"line_two"`
	UnitNumber   string    `gorm:"size:255;" json:"unit_number"`
	BusinessName   string    `gorm:"size:255;" json:"business_name"`
	AttentionTo   string    `gorm:"size:255;" json:"attention_to"`
	City     string    `gorm:"size:255;not null;" json:"city"`
	State     string    `gorm:"size:255;not null;" json:"state"`
	ZipCode     string    `gorm:"size:255;not null;" json:"zip_code"`
	Country     string    `gorm:"size:255;not null;" json:"country"`
	Phone   string    `gorm:"size:255;" json:"phone"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (a *Address) Prepare() {
	a.ID = 0
	a.Nickname = html.EscapeString(strings.TrimSpace(a.Nickname))
	a.LineOne = html.EscapeString(strings.TrimSpace(a.LineOne))
	a.LineTwo = html.EscapeString(strings.TrimSpace(a.LineTwo))
	a.UnitNumber = html.EscapeString(strings.TrimSpace(a.UnitNumber))
	a.BusinessName = html.EscapeString(strings.TrimSpace(a.BusinessName))
	a.AttentionTo = html.EscapeString(strings.TrimSpace(a.AttentionTo))
	a.City = html.EscapeString(strings.TrimSpace(a.City))
	a.State = html.EscapeString(strings.TrimSpace(a.State))
	a.ZipCode = html.EscapeString(strings.TrimSpace(a.ZipCode))
	a.Country = html.EscapeString(strings.TrimSpace(a.Country)
	a.Phone = html.EscapeString(strings.TrimSpace(a.Phone)
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
}

func (a *Address) Validate() error {

	if a.LineOne == "" {
		return errors.New("Address Line 1 Required")
	}
	if a.City == "" {
		return errors.New("City Required")
	}
	if a.State == "" {
		return errors.New("State Required")
	}
	if a.ZipCode == "" {
		return errors.New("Zip Code Required")
	}
	if a.Country == "" {
		return errors.New("Country Required")
	}
	return nil
}

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

func (p *Address) FindAllPosts(db *gorm.DB) (*[]Post, error) {
	var err error
	posts := []Post{}
	err = db.Debug().Model(&Post{}).Limit(100).Find(&posts).Error
	if err != nil {
		return &[]Post{}, err
	}
	if len(posts) > 0 {
		for i, _ := range posts {
			err := db.Debug().Model(&User{}).Where("id = ?", posts[i].AuthorID).Take(&posts[i].Author).Error
			if err != nil {
				return &[]Post{}, err
			}
		}
	}
	return &posts, nil
}

func (a *Address) FindAddressByID(db *gorm.DB, pid uint64) (*Address, error) {
	var err error
	err = db.Debug().Model(&Address{}).Where("id = ?", pid).Take(&a).Error
	if err != nil {
		return &Address{}, err
	}
	return a, nil
}

func (a *Address) UpdateAddress(db *gorm.DB) (*Address, error) {

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
	
	
	// Need to change which values to update
	err = db.Debug().Model(&Address{}).Where("id = ?", a.ID).Updates(Address{Title: a.Title, Content: a.Content, UpdatedAt: time.Now()}).Error
	if err != nil {
		return &Address{}, err
	}
	return a, nil
}

func (a *Address) DeleteAddress(db *gorm.DB, aid uint64) (int64, error) {

	db = db.Debug().Model(&Address{}).Where("id = ?", aid).Take(&Address{}).Delete(&Address{})

	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Address not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
