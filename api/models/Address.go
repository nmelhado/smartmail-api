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
	User    User      `json:"user"`
	UserID  uint32    `sql:"type:int REFERENCES users(id)" json:"user_id"`
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
	a.User = User{}
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
}

func (p *Address) Validate() error {

	if p.Title == "" {
		return errors.New("Required Title")
	}
	if p.Content == "" {
		return errors.New("Required Content")
	}
	if p.AuthorID < 1 {
		return errors.New("Required Author")
	}
	return nil
}

func (p *Address) SavePost(db *gorm.DB) (*Post, error) {
	var err error
	err = db.Debug().Model(&Post{}).Create(&p).Error
	if err != nil {
		return &Post{}, err
	}
	if p.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", p.AuthorID).Take(&p.Author).Error
		if err != nil {
			return &Post{}, err
		}
	}
	return p, nil
}

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

func (p *Address) FindPostByID(db *gorm.DB, pid uint64) (*Post, error) {
	var err error
	err = db.Debug().Model(&Post{}).Where("id = ?", pid).Take(&p).Error
	if err != nil {
		return &Post{}, err
	}
	if p.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", p.AuthorID).Take(&p.Author).Error
		if err != nil {
			return &Post{}, err
		}
	}
	return p, nil
}

func (p *Address) UpdateAPost(db *gorm.DB) (*Post, error) {

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
	err = db.Debug().Model(&Post{}).Where("id = ?", p.ID).Updates(Post{Title: p.Title, Content: p.Content, UpdatedAt: time.Now()}).Error
	if err != nil {
		return &Post{}, err
	}
	if p.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", p.AuthorID).Take(&p.Author).Error
		if err != nil {
			return &Post{}, err
		}
	}
	return p, nil
}

func (p *Address) DeleteAPost(db *gorm.DB, pid uint64, uid uint32) (int64, error) {

	db = db.Debug().Model(&Post{}).Where("id = ? and author_id = ?", pid, uid).Take(&Post{}).Delete(&Post{})

	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Post not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
