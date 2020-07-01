package responses

import (
	"sort"
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

// TokenResponse is the struct returned when a token request
type TokenResponse struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

// PasswordResetRequest is the struct returned when a user requests a password reset
type PasswordResetRequest struct {
	Token string `json:"token"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// PasswordReset is the struct returned when a user resets their password
type PasswordReset struct {
	Success bool   `json:"success"`
	Name    string `json:"name"`
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
	City                 string      `json:"city"`
	State                string      `json:"state"`
	ZipCode              string      `json:"zip_code"`
	Country              string      `json:"country"`
	Phone                null.String `json:"phone,omitempty"`
	DeliveryInstructions string      `json:"delivery_instructions,omitempty"`
}

// ZipResponse is to return a zip code to a retailer or mailer
type ZipResponse struct {
	SmartID string `json:"smart_id"`
	ZipCode string `json:"zip_code"`
}

// Contacts is the array of contacts response for a contact request
type Contacts struct {
	Contacts []Contact `json:"contacts"`
}

// Contact is the single contact response for a contact request
type Contact struct {
	Name    string    `json:"name"`
	SmartID string    `json:"smart_id"`
	Phone   string    `json:"phone"`
	Email   string    `json:"email"`
	AddedOn time.Time `json:"added_on"`
}

// PackagesResponse is the array of open packages and delivered packages response for a packages request
type PackagesResponse struct {
	OpenPackages      []SinglePackage `json:"open_packages"`
	DeliveredPackages []SinglePackage `json:"delivered_packages"`
	Success           bool            `json:"success"`
}

// SinglePackage is the single package response for a package request
type SinglePackage struct {
	PackageID          uint64             `json:"id"` // actually the package description ID
	MailCarrier        string             `json:"mail_carrier"`
	Sender             SenderRecipient    `json:"sender"`
	Recipient          SenderRecipient    `json:"recipient"`
	Tracking           string             `json:"tracking"`
	EstimatedDelivery  null.Time          `json:"estimatedDelivery"`
	DeliveredOn        null.Time          `json:"delivered_on"`
	PackageDescription PackageDescription `json:"package_description"`
}

// SenderRecipient contains information about the sender or recipient
type SenderRecipient struct {
	Name        null.String `json:"name"`
	SmartID     null.String `json:"smart_id"`
	LargeLogo   null.String `json:"large_logo"`
	SmallLogo   null.String `json:"small_logo"`
	RedirectURL null.String `json:"redirect_url"`
	Role        null.String `json:"role"`
}

// PackageDescription contains additional package information for a package request
type PackageDescription struct {
	Contents   null.String `json:"contents"`
	OrderLink  null.String `json:"order_link"`
	OrderImage null.String `json:"order_image"`
}

// UpdatePackageResponse is the response for an update package request
type UpdatePackageResponse struct {
	Success   bool   `json:"success"`
	Tracking  string `json:"tracking"`
	Delivered bool   `json:"delivered"`
}

// UpdatePackageDescriptionResponse is the response for an update package request
type UpdatePackageDescriptionResponse struct {
	Success    bool        `json:"success"`
	Contents   null.String `json:"contents"`
	OrderLink  null.String `json:"order_link"`
	OrderImage null.String `json:"order_image"`
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

// TranslateSmartAddressResponse converts an AddressAssignments into an AddressSmartIDResponse
func TranslateSmartAddressResponse(originalAddress *models.AddressAssignment, reply *AddressSmartIDResponse) {
	reply.SmartID = originalAddress.User.SmartID
	reply.FirstName = originalAddress.User.FirstName
	reply.LastName = originalAddress.User.LastName
	reply.BusinessName = originalAddress.Address.BusinessName.String
	reply.AttentionTo = originalAddress.Address.AttentionTo.String
	reply.LineOne = originalAddress.Address.LineOne
	reply.LineTwo = originalAddress.Address.LineTwo.String
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

// TranslateZipResponse converts an AddressAssignment into a ZipResponse
func TranslateZipResponse(originalAddress *models.AddressAssignment, reply *ZipResponse) {
	reply.SmartID = originalAddress.User.SmartID
	reply.ZipCode = originalAddress.Address.ZipCode
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
		nextContact := TranslateContact(contact)
		contacts = append(contacts, nextContact)
	}

	// orders by first name and then last name
	sort.SliceStable(contacts, func(i, j int) bool { return contacts[i].Name < contacts[j].Name })
	return
}

// TranslateContact converts a single contact into a contact response
func TranslateContact(originalContact models.Contact) (contact Contact) {
	contact.Name = originalContact.Contact.FirstName + " " + originalContact.Contact.LastName
	contact.SmartID = originalContact.Contact.SmartID
	contact.Email = originalContact.Contact.Email
	contact.Phone = originalContact.Contact.Phone
	contact.AddedOn = originalContact.CreatedAt
	return
}

// TranslatePackagesResponse converts an array of packages into an array of package responses
func TranslatePackagesResponse(openPackages []models.Package, deliveredPackages []models.Package) (userPackages PackagesResponse) {
	// translate open packages
	for _, singlePackage := range openPackages {
		nextPackage := TranslatePackage(singlePackage)
		userPackages.OpenPackages = append(userPackages.OpenPackages, nextPackage)
	}

	// translate delivered packages
	for _, singlePackage := range deliveredPackages {
		nextPackage := TranslatePackage(singlePackage)
		userPackages.DeliveredPackages = append(userPackages.DeliveredPackages, nextPackage)
	}

	userPackages.Success = true

	return
}

// TranslatePackage converts a single package into a package response
func TranslatePackage(originalPackage models.Package) (newPackage SinglePackage) {
	newPackage.MailCarrier = originalPackage.MailCarrier.Name
	if originalPackage.SenderID.Valid {
		newPackage.Sender = SenderRecipient{
			Name:        null.StringFrom(originalPackage.Sender.FirstName + " " + originalPackage.Sender.LastName),
			SmartID:     null.StringFrom(originalPackage.Sender.SmartID),
			LargeLogo:   originalPackage.Sender.LargeLogo,
			SmallLogo:   originalPackage.Sender.SmallLogo,
			RedirectURL: originalPackage.Sender.RedirectURL,
			Role:        null.StringFrom(string(originalPackage.Sender.Authority)),
		}
	}
	if originalPackage.RecipientID.Valid {
		newPackage.Recipient = SenderRecipient{
			Name:        null.StringFrom(originalPackage.Recipient.FirstName + " " + originalPackage.Recipient.LastName),
			SmartID:     null.StringFrom(originalPackage.Recipient.SmartID),
			LargeLogo:   originalPackage.Recipient.LargeLogo,
			SmallLogo:   originalPackage.Recipient.SmallLogo,
			RedirectURL: originalPackage.Recipient.RedirectURL,
			Role:        null.StringFrom(string(originalPackage.Recipient.Authority)),
		}
	}
	newPackage.Tracking = originalPackage.Tracking.String
	if originalPackage.Delivered {
		newPackage.DeliveredOn = originalPackage.DeliveredOn
	}
	newPackage.EstimatedDelivery = originalPackage.EstimatedDelivery
	if originalPackage.PackageDescriptionID > 0 {
		newPackage.PackageDescription = PackageDescription{
			Contents:   originalPackage.PackageDescription.Contents,
			OrderImage: originalPackage.PackageDescription.OrderImage,
			OrderLink:  originalPackage.PackageDescription.OrderLink,
		}
	}
	return
}
