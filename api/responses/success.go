package responses

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

// Struct returned when a new user is created
type CreateUserResponse struct {
	CosmoID   string    `json:"cosmo_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

// Struct returned when a new user and address are simultaneously created
type CreateUserAndAddressResponse struct {
	CosmoID      string      `json:"cosmo_id"`
	Email        string      `json:"email"`
	FirstName    string      `json:"first_name"`
	LastName     string      `json:"last_name"`
	Phone        string      `json:"phone"`
	BusinessName null.String `json:"business_name"`
	AttentionTo  null.String `json:"attention_to"`
	LineOne      string      `json:"line_one"`
	LineTwo      null.String `json:"line_two"`
	UnitNumber   null.String `json:"unit_number"`
	City         string      `json:"city"`
	State        string      `json:"state"`
	ZipCode      string      `json:"zip_code"`
	Country      string      `json:"country"`
	CreatedAt    time.Time   `json:"created_at"`
	AddressPhone null.String `json:"phone_for_address"`
}

// Used for creating, updating, and retrieving a single address
type AddressResponse struct {
	FirstName    string      `json:"first_name"`
	LastName     string      `json:"last_name"`
	BusinessName null.String `json:"business_name"`
	AttentionTo  null.String `json:"attention_to"`
	LineOne      string      `json:"line_one"`
	LineTwo      null.String `json:"line_two"`
	UnitNumber   null.String `json:"unit_number"`
	City         string      `json:"city"`
	State        string      `json:"state"`
	ZipCode      string      `json:"zip_code"`
	Country      string      `json:"country"`
	Phone        null.String `json:"phone"`
}
