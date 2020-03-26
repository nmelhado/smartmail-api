package responses

import (
	"time"

	"github.com/nmelhado/smartmail-api/api/models"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
)

// CreateUserResponse is the struct returned when a new user is created
type CreateUserResponse struct {
	ID        uuid.UUID `json:"id"`
	SmartID   string    `json:"smart_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

// UserAndAddressResponse is the struct returned when a new user and address are simultaneously created
type UserAndAddressResponse struct {
	User      CreateUserResponse `json:"user"`
	Addresses []BasicAddress     `json:"addresses"`
	Token     string             `json:"token"`
	Expires   time.Time          `json:"expires"`
}

// BasicAddress is used to return an address to the UI
type BasicAddress struct {
	Status       models.Status `json:"address_type"`
	StartDate    time.Time     `json:"start_date"`
	EndDate      string        `json:"end_date,omitempty"`
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

// AddressResponse is used for creating, updating, and retrieving a single address
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

// AddressSmartIDResponse is used for creating, updating, and retrieving a single address
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

// TranslateAddressResponse converts an array of AddressAssignments into an array of AddressResponse
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

// TranslateSmartAddressResponse converts an array of AddressAssignments into an array of AddressSmartIDResponse
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

// TranslateUserAndAddressResponse converts an AddressAssignment into a UserAndAddressResponse
func TranslateUserAndAddressResponse(originalAddress *models.AddressAssignment, reply *UserAndAddressResponse) {
	reply.User.ID = originalAddress.User.ID
	reply.User.SmartID = originalAddress.User.SmartID
	reply.User.Email = originalAddress.User.Email
	reply.User.Phone = originalAddress.User.Phone
	reply.User.FirstName = originalAddress.User.FirstName
	reply.User.LastName = originalAddress.User.LastName
	reply.User.CreatedAt = originalAddress.User.CreatedAt

	replyAddress := BasicAddress{
		Status:       originalAddress.Status,
		StartDate:    originalAddress.StartDate,
		BusinessName: originalAddress.Address.BusinessName.String,
		AttentionTo:  originalAddress.Address.AttentionTo.String,
		LineOne:      originalAddress.Address.LineOne,
		LineTwo:      originalAddress.Address.LineTwo.String,
		UnitNumber:   originalAddress.Address.UnitNumber.String,
		City:         originalAddress.Address.City,
		State:        originalAddress.Address.State,
		ZipCode:      originalAddress.Address.ZipCode,
		Country:      originalAddress.Address.Country,
		Phone:        originalAddress.Address.Phone,
	}
	if originalAddress.EndDate.Valid {
		replyAddress.EndDate = originalAddress.EndDate.Time.String()
	}
	if !replyAddress.Phone.Valid {
		replyAddress.Phone.SetValid(originalAddress.User.Phone)
	}
	reply.Addresses = []BasicAddress{replyAddress}
}

// TranslateAddresses converts an array of AddressAssignments into an array of BasicAddresses
func TranslateAddresses(originalAddresses *[]models.AddressAssignment) (addresses []BasicAddress) {
	for _, address := range *originalAddresses {
		nextAddress := BasicAddress{
			Status:       address.Status,
			StartDate:    address.StartDate,
			BusinessName: address.Address.BusinessName.String,
			AttentionTo:  address.Address.AttentionTo.String,
			LineOne:      address.Address.LineOne,
			LineTwo:      address.Address.LineTwo.String,
			UnitNumber:   address.Address.UnitNumber.String,
			City:         address.Address.City,
			State:        address.Address.State,
			ZipCode:      address.Address.ZipCode,
			Country:      address.Address.Country,
			Phone:        address.Address.Phone,
		}
		if address.EndDate.Valid {
			nextAddress.EndDate = address.EndDate.Time.String()
		}
		if !nextAddress.Phone.Valid {
			nextAddress.Phone.SetValid(address.User.Phone)
		}
		addresses = append(addresses, nextAddress)
	}
	return
}
