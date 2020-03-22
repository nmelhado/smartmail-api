package responses

import (
	"time"

	"github.com/nmelhado/smartmail-api/api/models"
	"gopkg.in/guregu/null.v3"
)

// Struct returned when logging in
type TokenResponse struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

// Struct returned when a new user is created
type CreateUserResponse struct {
	SmartID   string    `json:"smart_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

// Struct returned when a new user and address are simultaneously created
type CreateUserAndAddressResponse struct {
	SmartID      string    `json:"smart_id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Phone        string    `json:"phone"`
	BusinessName string    `json:"business_name,omitempty"`
	AttentionTo  string    `json:"attention_to,omitempty"`
	LineOne      string    `json:"line_one"`
	LineTwo      string    `json:"line_two,omitempty"`
	UnitNumber   string    `json:"unit_number,omitempty"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	ZipCode      string    `json:"zip_code"`
	Country      string    `json:"country"`
	AddressPhone string    `json:"phone_for_address,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Used for creating, updating, and retrieving a single address
type AddressResponse struct {
	Status       models.Status `json:"address_type"`
	StartDate    time.Time     `json:"start_date"`
	EndDate      string        `json:"end_date,omitempty"`
	FirstName    string        `json:"first_name"`
	LastName     string        `json:"last_name"`
	BusinessName string        `json:"business_name,omitempty"`
	AttentionTo  string        `json:"attention_to,omitempty"`
	LineOne      string        `json:"line_one"`
	LineTwo      string        `json:"line_two,omitempty"`
	UnitNumber   string        `json:"unit_number,omitempty"`
	City         string        `json:"city"`
	State        string        `json:"state"`
	ZipCode      string        `json:"zip_code"`
	Country      string        `json:"country"`
	Phone        null.String   `json:"phone,omitempty"`
}

// Used for creating, updating, and retrieving a single address
type AddressSmartIDResponse struct {
	SmartID      string      `json:"smart_id"`
	FirstName    string      `json:"first_name"`
	LastName     string      `json:"last_name"`
	BusinessName string      `json:"business_name,omitempty"`
	AttentionTo  string      `json:"attention_to,omitempty"`
	LineOne      string      `json:"line_one"`
	LineTwo      string      `json:"line_two,omitempty"`
	UnitNumber   string      `json:"unit_number,omitempty"`
	City         string      `json:"city"`
	State        string      `json:"state"`
	ZipCode      string      `json:"zip_code"`
	Country      string      `json:"country"`
	Phone        null.String `json:"phone,omitempty"`
}

func TranslateAddressResponse(originalAddress *models.AddressAssignment, reply *AddressResponse) {
	reply.Status = originalAddress.Status
	reply.StartDate = originalAddress.StartDate
	if originalAddress.EndDate.Valid {
		reply.EndDate = originalAddress.EndDate.Time.String()
	}
	reply.FirstName = originalAddress.User.FirstName
	reply.LastName = originalAddress.User.LastName
	reply.BusinessName = originalAddress.Address.BusinessName.String
	reply.AttentionTo = originalAddress.Address.AttentionTo.String
	reply.LineOne = originalAddress.Address.LineOne
	reply.LineTwo = originalAddress.Address.LineTwo.String
	reply.UnitNumber = originalAddress.Address.UnitNumber.String
	reply.City = originalAddress.Address.City
	reply.State = originalAddress.Address.State
	reply.ZipCode = originalAddress.Address.ZipCode
	reply.Country = originalAddress.Address.Country
	reply.Phone = originalAddress.Address.Phone
	if !reply.Phone.Valid {
		reply.Phone.SetValid(originalAddress.User.Phone)
	}
}

func TranslateSmartAddressResponse(originalAddress *models.AddressAssignment, reply *AddressSmartIDResponse) {
	reply.SmartID = originalAddress.User.SmartID
	reply.FirstName = originalAddress.User.FirstName
	reply.LastName = originalAddress.User.LastName
	reply.BusinessName = originalAddress.Address.BusinessName.String
	reply.AttentionTo = originalAddress.Address.AttentionTo.String
	reply.LineOne = originalAddress.Address.LineOne
	reply.LineTwo = originalAddress.Address.LineTwo.String
	reply.UnitNumber = originalAddress.Address.UnitNumber.String
	reply.City = originalAddress.Address.City
	reply.State = originalAddress.Address.State
	reply.ZipCode = originalAddress.Address.ZipCode
	reply.Country = originalAddress.Address.Country
	reply.Phone = originalAddress.Address.Phone
	if !reply.Phone.Valid {
		reply.Phone.SetValid(originalAddress.User.Phone)
	}
}

func TranslateUserAndAddressResponse(originalAddress *models.AddressAssignment, reply *CreateUserAndAddressResponse) {
	reply.SmartID = originalAddress.User.SmartID
	reply.Email = originalAddress.User.Email
	reply.Phone = originalAddress.User.Phone
	reply.FirstName = originalAddress.User.FirstName
	reply.LastName = originalAddress.User.LastName
	reply.BusinessName = originalAddress.Address.BusinessName.String
	reply.AttentionTo = originalAddress.Address.AttentionTo.String
	reply.LineOne = originalAddress.Address.LineOne
	reply.LineTwo = originalAddress.Address.LineTwo.String
	reply.UnitNumber = originalAddress.Address.UnitNumber.String
	reply.City = originalAddress.Address.City
	reply.State = originalAddress.Address.State
	reply.ZipCode = originalAddress.Address.ZipCode
	reply.Country = originalAddress.Address.Country
	reply.AddressPhone = originalAddress.Address.Phone.String
	reply.CreatedAt = originalAddress.User.CreatedAt
}
