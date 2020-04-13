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
	Contacts  []Contact          `json:"contacts"`
	Token     string             `json:"token"`
	Expires   time.Time          `json:"expires"`
}

// AddressesResponse is the struct returned when a new user and address are simultaneously created
type AddressesResponse struct {
	Addresses []BasicAddress `json:"addresses"`
}

// BasicAddress is used to return an address to the UI
type BasicAddress struct {
	ID                   uint64        `json:"id"`
	Nickname             null.String   `json:"nickname"`
	Status               models.Status `json:"address_type"`
	StartDate            time.Time     `json:"start_date"`
	EndDate              string        `json:"end_date,omitempty"`
	BusinessName         string        `json:"business_name,omitempty"`
	AttentionTo          string        `json:"attention_to,omitempty"`
	LineOne              string        `json:"line_one"`
	LineTwo              string        `json:"line_two,omitempty"`
	UnitNumber           string        `json:"unit_number,omitempty"`
	City                 string        `json:"city"`
	State                string        `json:"state"`
	ZipCode              string        `json:"zip_code"`
	Country              string        `json:"country"`
	Phone                null.String   `json:"phone,omitempty"`
	Latitude             float64       `json:"latitude"`
	Longitude            float64       `json:"longitude"`
	DeliveryInstructions string        `json:"delivery_instructions,omitempty"`
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

// ToAndFromAddressSmartIDResponse is used to return sender and recipeint addresses to a mailer
type ToAndFromAddressSmartIDResponse struct {
	Sender    AddressSmartIDResponse `json:"sender"`
	Recipient AddressSmartIDResponse `json:"recipient"`
}

// AddressSmartIDResponse is used for creating, updating, and retrieving a single address
type AddressSmartIDResponse struct {
	SmartID              string      `json:"smart_id"`
	FirstName            string      `json:"first_name"`
	LastName             string      `json:"last_name"`
	BusinessName         string      `json:"business_name,omitempty"`
	AttentionTo          string      `json:"attention_to,omitempty"`
	LineOne              string      `json:"line_one"`
	LineTwo              string      `json:"line_two,omitempty"`
	UnitNumber           string      `json:"unit_number,omitempty"`
	City                 string      `json:"city"`
	State                string      `json:"state"`
	ZipCode              string      `json:"zip_code"`
	Country              string      `json:"country"`
	Phone                null.String `json:"phone,omitempty"`
	DeliveryInstructions string      `json:"delivery_instructions,omitempty"`
}

// Contacts is the array of contacts response for a contact request
type Contacts struct {
	Contacts []Contact `json:"contacts"`
}

// Contact is the single contact response for a contact request
type Contact struct {
	Name    string `json:"name"`
	SmartID string `json:"smart_id"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
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

// TranslateToAndFromSmartAddressResponse converts sender and recipient AddressAssignments into a ToAndFromAddressSmartIDResponse
func TranslateToAndFromSmartAddressResponse(senderAddress *models.AddressAssignment, recipientAddress *models.AddressAssignment, reply *ToAndFromAddressSmartIDResponse) {
	TranslateSmartAddressResponse(senderAddress, &reply.Sender)
	TranslateSmartAddressResponse(recipientAddress, &reply.Recipient)
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
	if !reply.Phone.Valid || reply.Phone.String == "" {
		reply.Phone.SetValid(originalAddress.User.Phone)
	}
	reply.DeliveryInstructions = originalAddress.Address.DeliveryInstructions.String
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
		ID:                   originalAddress.ID,
		Nickname:             originalAddress.Address.Nickname,
		Status:               originalAddress.Status,
		StartDate:            originalAddress.StartDate,
		BusinessName:         originalAddress.Address.BusinessName.String,
		AttentionTo:          originalAddress.Address.AttentionTo.String,
		LineOne:              originalAddress.Address.LineOne,
		LineTwo:              originalAddress.Address.LineTwo.String,
		UnitNumber:           originalAddress.Address.UnitNumber.String,
		City:                 originalAddress.Address.City,
		State:                originalAddress.Address.State,
		ZipCode:              originalAddress.Address.ZipCode,
		Country:              originalAddress.Address.Country,
		Phone:                originalAddress.Address.Phone,
		Latitude:             originalAddress.Address.Latitude,
		Longitude:            originalAddress.Address.Longitude,
		DeliveryInstructions: originalAddress.Address.DeliveryInstructions.String,
	}
	if originalAddress.EndDate.Valid {
		replyAddress.EndDate = originalAddress.EndDate.Time.String()
	}
	if !replyAddress.Phone.Valid || replyAddress.Phone.String == "" {
		replyAddress.Phone.SetValid(originalAddress.User.Phone)
	}
	reply.Addresses = []BasicAddress{replyAddress}
}

// TranslateAddress converts a single AddressAssignment into a BasicAddresses
func TranslateAddress(originalAddress *models.AddressAssignment) (address BasicAddress) {
	address = BasicAddress{
		ID:                   originalAddress.ID,
		Nickname:             originalAddress.Address.Nickname,
		Status:               originalAddress.Status,
		StartDate:            originalAddress.StartDate,
		BusinessName:         originalAddress.Address.BusinessName.String,
		AttentionTo:          originalAddress.Address.AttentionTo.String,
		LineOne:              originalAddress.Address.LineOne,
		LineTwo:              originalAddress.Address.LineTwo.String,
		UnitNumber:           originalAddress.Address.UnitNumber.String,
		City:                 originalAddress.Address.City,
		State:                originalAddress.Address.State,
		ZipCode:              originalAddress.Address.ZipCode,
		Country:              originalAddress.Address.Country,
		Phone:                originalAddress.Address.Phone,
		Latitude:             originalAddress.Address.Latitude,
		Longitude:            originalAddress.Address.Longitude,
		DeliveryInstructions: originalAddress.Address.DeliveryInstructions.String,
	}
	if originalAddress.EndDate.Valid {
		address.EndDate = originalAddress.EndDate.Time.String()
	}
	if !address.Phone.Valid || address.Phone.String == "" {
		address.Phone.SetValid(originalAddress.User.Phone)
	}
	return
}

// TranslateAddresses converts an array of AddressAssignments into an array of BasicAddresses
func TranslateAddresses(originalAddresses *[]models.AddressAssignment) (addresses []BasicAddress) {
	for _, address := range *originalAddresses {
		nextAddress := TranslateAddress(&address)
		addresses = append(addresses, nextAddress)
	}
	return
}

// TranslateContacts converts an array of contacts into an array of contacts response
func TranslateContacts(originalContacts []models.Contact) (contacts []Contact) {
	for _, contact := range originalContacts {
		nextContact := translateContact(contact)
		contacts = append(contacts, nextContact)
	}
	return
}

// TranslateContact converts a single contact into a contact response
func translateContact(originalContact models.Contact) (contact Contact) {
	contact.Name = originalContact.Contact.FirstName + " " + originalContact.Contact.LastName
	contact.SmartID = originalContact.Contact.SmartID
	contact.Email = originalContact.Contact.Email
	contact.Phone = originalContact.Contact.Phone
	return
}
