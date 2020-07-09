package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/nmelhado/smartmail-api/api/auth"
	"github.com/nmelhado/smartmail-api/api/models"
	"github.com/nmelhado/smartmail-api/api/responses"
	"github.com/nmelhado/smartmail-api/api/utils/formaterror"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
)

type geoInfo struct {
	Results      []result `json:"results"`
	ErrorMessage string   `json:"error_message"`
}

type result struct {
	Geometry geometry `json:"geometry"`
}

type geometry struct {
	Location latLng `json:"location"`
}

type latLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// CreateAddress creates an address and uploads the address to the DB and links it to the user that created it
func (server *Server) CreateAddress(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	addressAssignment := models.AddressAssignment{}
	err = json.Unmarshal(body, &addressAssignment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	addressAssignment.Address.Prepare()
	err = addressAssignment.Address.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	uid, err := auth.ExtractUITokenID(r)
	if err != nil {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	if uid != addressAssignment.UserID {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	user := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("id = ?", uid).Take(&user).Error
	if err != nil {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	addressAssignment.User = user

	err = geoLocate(&addressAssignment)
	if err != nil {
		_, _ = addressAssignment.User.DeleteUser(server.DB, user.ID)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	createAddress, err := addressAssignment.Address.SaveAddress(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	addressAssignment.AddressID = createAddress.ID

	addressAssignment.Prepare()
	err = addressAssignment.Validate()
	if err != nil {
		_, _ = addressAssignment.Address.DeleteAddress(server.DB, createAddress.ID)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	finalAddress := &models.AddressAssignment{}
	finalAddress, err = addressAssignment.SaveAddressAssignment(server.DB)
	if err != nil {
		_, _ = addressAssignment.Address.DeleteAddress(server.DB, createAddress.ID)
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	addressResponse := responses.TranslateAddress(finalAddress)

	type addressResponseStruct struct {
		Address responses.BasicAddress `json:"address"`
	}

	finalResponse := &addressResponseStruct{addressResponse}

	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, createAddress.ID))
	responses.JSON(w, http.StatusCreated, finalResponse)
}

// CreateUserAndAddress creates a user and address simultaneously
func (server *Server) CreateUserAndAddress(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	addressAssignment := models.AddressAssignment{}
	err = json.Unmarshal(body, &addressAssignment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	password := addressAssignment.User.Password
	addressAssignment.User.Prepare()
	err = addressAssignment.User.Validate("create")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user, err := addressAssignment.User.SaveUser(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	addressAssignment.Address.Prepare()
	err = addressAssignment.Address.Validate()
	if err != nil {
		_, _ = addressAssignment.User.DeleteUser(server.DB, user.ID)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	err = geoLocate(&addressAssignment)
	if err != nil {
		_, _ = addressAssignment.User.DeleteUser(server.DB, user.ID)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	addressAssignment.UserID = user.ID

	createAddress, err := addressAssignment.Address.SaveAddress(server.DB)
	if err != nil {
		_, _ = addressAssignment.User.DeleteUser(server.DB, user.ID)
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	addressAssignment.AddressID = createAddress.ID

	addressAssignment.Prepare()
	err = addressAssignment.Validate()
	if err != nil {
		_, _ = addressAssignment.User.DeleteUser(server.DB, user.ID)
		_, _ = addressAssignment.Address.DeleteAddress(server.DB, createAddress.ID)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	finalAddress := &models.AddressAssignment{}
	finalAddress, err = addressAssignment.SaveAddressAssignment(server.DB)
	if err != nil {
		_, _ = addressAssignment.User.DeleteUser(server.DB, user.ID)
		_, _ = addressAssignment.Address.DeleteAddress(server.DB, createAddress.ID)
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	token, _, err := server.SignIn(user.Email, password)
	if err != nil {
		_, _ = addressAssignment.User.DeleteUser(server.DB, user.ID)
		_, _ = addressAssignment.Address.DeleteAddress(server.DB, createAddress.ID)
		_, _ = addressAssignment.DeleteAddressAssignment(server.DB, finalAddress.ID)
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}

	addressResponse := &responses.UserAndAddressResponse{}
	responses.TranslateUserAndAddressResponse(finalAddress, addressResponse)

	addressResponse.Token = token
	addressResponse.Expires = time.Now().Add(time.Hour * 1)

	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, createAddress.ID))
	responses.JSON(w, http.StatusCreated, addressResponse)
}

// GetAddressByID retrieves an address using an address ID
func (server *Server) GetAddressByID(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	aid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	address := models.Address{}

	addressReceived, err := address.FindAddressByID(server.DB, aid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, addressReceived)
}

// GetMailingAddressToAndFromBySmartID retrieves a user's mailing address using a customer's SmartID
func (server *Server) GetMailingAddressToAndFromBySmartID(w http.ResponseWriter, r *http.Request) {
	reqUID, permission, err := auth.ExtractAPIUserTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	reqUser := models.APIUser{}

	err = server.DB.Debug().Model(models.APIUser{}).Where("id = ?", reqUID).Take(&reqUser).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	hasPermission := models.FullAddressPermissions(reqUser.Permission)

	if !hasPermission || string(reqUser.Permission) != permission {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	vars := mux.Vars(r)
	senderSmartID := sanitizeSmartID(strings.ToUpper(vars["sender_smart_id"]))
	recipientSmartID := sanitizeSmartID(strings.ToUpper(vars["recipient_smart_id"]))
	date, err := time.Parse("2006-01-02", vars["date"])
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	sender := models.User{}
	recipient := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("smart_id = ?", senderSmartID).Take(&sender).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("Unable to find sender with smartID: %s", senderSmartID))
		return
	}

	err = server.DB.Debug().Model(models.User{}).Where("smart_id = ?", recipientSmartID).Take(&recipient).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("Unable to find recipient with smartID: %s", recipientSmartID))
		return
	}

	senderAddress := models.AddressAssignment{}
	senderAddressReceived, err := senderAddress.FindMailingAddressWithSmartID(server.DB, sender, date)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	recipientAddress := models.AddressAssignment{}
	recipientAddressReceived, err := recipientAddress.FindMailingAddressWithSmartID(server.DB, recipient, date)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	contacts := models.Contact{}
	err = contacts.SaveContacts(server.DB, sender.ID, recipient.ID)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	addressResponse := &responses.ToAndFromAddressSmartIDResponse{}
	responses.TranslateToAndFromSmartAddressResponse(senderAddressReceived, recipientAddressReceived, addressResponse)
	addressResponse.Sender.DeliveryInstructions = ""
	addressResponse.Recipient.DeliveryInstructions = ""

	responses.JSON(w, http.StatusOK, addressResponse)
}

// GetMailingAddressBySmartID retrieves a user's mailing address using a customer's SmartID
func (server *Server) GetMailingAddressBySmartID(w http.ResponseWriter, r *http.Request) {
	reqUID, permission, err := auth.ExtractAPIUserTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	reqUser := models.APIUser{}

	err = server.DB.Debug().Model(models.APIUser{}).Where("id = ?", reqUID).Take(&reqUser).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	hasPermission := models.FullAddressPermissions(reqUser.Permission)

	if !hasPermission || string(reqUser.Permission) != permission {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	vars := mux.Vars(r)
	smartID := sanitizeSmartID(strings.ToUpper(vars["smart_id"]))
	date, err := time.Parse("2006-01-02", vars["date"])
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	user := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("smart_id = ?", smartID).Take(&user).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("Unable to find smartID: %s", smartID))
		return
	}

	address := models.AddressAssignment{}
	addressReceived, err := address.FindMailingAddressWithSmartID(server.DB, user, date)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	addressResponse := &responses.AddressSmartIDResponse{}
	responses.TranslateSmartAddressResponse(addressReceived, addressResponse)
	addressResponse.DeliveryInstructions = ""

	responses.JSON(w, http.StatusOK, addressResponse)
}

// GetMailingZipBySmartID retrieves a user's zip code using a customer's SmartID
func (server *Server) GetMailingZipBySmartID(w http.ResponseWriter, r *http.Request) {
	reqUID, permission, err := auth.ExtractAPIUserTokenID(r)
	if err != nil {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	reqUser := models.APIUser{}

	err = server.DB.Debug().Model(models.APIUser{}).Where("id = ?", reqUID).Take(&reqUser).Error
	if err != nil {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	hasPermission := models.ZipPermissions(reqUser.Permission)

	fmt.Printf("permission: %s", permission)
	fmt.Printf("reqUser.Permission: %s", reqUser.Permission)
	fmt.Printf("hasPermission: %v", hasPermission)

	if !hasPermission || string(reqUser.Permission) != permission {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	vars := mux.Vars(r)
	smartID := sanitizeSmartID(strings.ToUpper(vars["smart_id"]))
	date, err := time.Parse("2006-01-02", vars["date"])
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	user := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("smart_id = ?", smartID).Take(&user).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("Unable to find smartID: %s", smartID))
		return
	}

	address := models.AddressAssignment{}
	addressReceived, err := address.FindMailingAddressWithSmartID(server.DB, user, date)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	zipResponse := &responses.ZipResponse{}
	responses.TranslateZipResponse(addressReceived, zipResponse)

	responses.JSON(w, http.StatusOK, zipResponse)
}

// GetPackageAddressToAndFromBySmartID retrieves a user's mailing address using a customer's SmartID
func (server *Server) GetPackageAddressToAndFromBySmartID(w http.ResponseWriter, r *http.Request) {
	reqUID, permission, err := auth.ExtractAPIUserTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	reqUser := models.APIUser{}

	err = server.DB.Debug().Model(models.APIUser{}).Where("id = ?", reqUID).Take(&reqUser).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	hasPermission := models.FullAddressPermissions(reqUser.Permission)

	if !hasPermission || string(reqUser.Permission) != permission {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	vars := mux.Vars(r)
	senderSmartID := sanitizeSmartID(strings.ToUpper(vars["sender_smart_id"]))
	recipientSmartID := sanitizeSmartID(strings.ToUpper(vars["recipient_smart_id"]))
	tracking := null.StringFromPtr(nil)
	trackingQuery, ok := vars["tracking"]
	if ok {
		tracking = null.StringFrom(trackingQuery)
	}
	date, err := time.Parse("2006-01-02", vars["date"])
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	sender := models.User{}
	recipient := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("smart_id = ?", senderSmartID).Take(&sender).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("Unable to find sender with smartID: %s", senderSmartID))
		return
	}

	err = server.DB.Debug().Model(models.User{}).Where("smart_id = ?", recipientSmartID).Take(&recipient).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("Unable to find recipient with smartID: %s", recipientSmartID))
		return
	}

	senderAddress := models.AddressAssignment{}
	senderAddressReceived, err := senderAddress.FindPackageAddressWithSmartID(server.DB, sender, date)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	recipientAddress := models.AddressAssignment{}
	recipientAddressReceived, err := recipientAddress.FindPackageAddressWithSmartID(server.DB, recipient, date)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	contacts := models.Contact{}
	err = contacts.SaveContacts(server.DB, sender.ID, recipient.ID)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	packageDescription := &models.PackageDescription{}
	packageDescription, err = packageDescription.SavePackageDescription(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	newPackage := models.Package{
		MailCarrierID: reqUID,
		SenderID: uuid.NullUUID{
			UUID:  sender.ID,
			Valid: true,
		},
		RecipientID: uuid.NullUUID{
			UUID:  recipient.ID,
			Valid: true,
		},
		Tracking:             tracking,
		PackageDescriptionID: packageDescription.ID,
	}
	err = newPackage.SavePackage(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	addressResponse := &responses.ToAndFromAddressSmartIDResponse{}
	responses.TranslateToAndFromSmartAddressResponse(senderAddressReceived, recipientAddressReceived, addressResponse)

	responses.JSON(w, http.StatusOK, addressResponse)
}

// AddressAndInfoRequest is the struct to receive a shipping info request
// with additional package description within a POST request from the mailer
type AddressAndInfoRequest struct {
	SenderSmartID      null.String               `json:"sender_smart_id"`
	RecipientSmartID   null.String               `json:"recipient_smart_id"`
	Tracking           string                    `json:"tracking"`
	TargetDate         string                    `json:"target_date"`
	Date               null.Time                 `json:"date"`
	PackageDescription models.PackageDescription `json:"package_description"`
}

// ProvidePackageAddressToAndFromBySmartID retrieves a user's mailing address using a customer's SmartID and sets package description
func (server *Server) ProvidePackageAddressToAndFromBySmartID(w http.ResponseWriter, r *http.Request) {
	reqUID, permission, err := auth.ExtractAPIUserTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	reqUser := models.APIUser{}

	err = server.DB.Debug().Model(models.APIUser{}).Where("id = ?", reqUID).Take(&reqUser).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	hasPermission := models.FullAddressPermissions(reqUser.Permission)

	if !hasPermission || string(reqUser.Permission) != permission {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	addressAndInfoRequest := AddressAndInfoRequest{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	err = json.Unmarshal(body, &addressAndInfoRequest)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	if !addressAndInfoRequest.SenderSmartID.Valid && !addressAndInfoRequest.RecipientSmartID.Valid {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("Either a sender smartID or a recipient smartID must be provided"))
		return
	}

	if addressAndInfoRequest.TargetDate == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("Please provide a valid date"))
		return
	}

	date, err := time.Parse("2006-01-02", addressAndInfoRequest.TargetDate)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	addressAndInfoRequest.Date = null.TimeFrom(date)

	sender := models.User{}
	recipient := models.User{}

	if addressAndInfoRequest.SenderSmartID.Valid {
		addressAndInfoRequest.SenderSmartID = null.StringFrom(sanitizeSmartID(strings.ToUpper(addressAndInfoRequest.SenderSmartID.String)))
		err = server.DB.Debug().Model(models.User{}).Where("smart_id = ?", addressAndInfoRequest.SenderSmartID.String).Take(&sender).Error
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("Unable to find sender with smartID: %s", addressAndInfoRequest.SenderSmartID.String))
			return
		}
	}

	if addressAndInfoRequest.RecipientSmartID.Valid {
		addressAndInfoRequest.RecipientSmartID = null.StringFrom(sanitizeSmartID(strings.ToUpper(addressAndInfoRequest.RecipientSmartID.String)))
		err = server.DB.Debug().Model(models.User{}).Where("smart_id = ?", addressAndInfoRequest.RecipientSmartID.String).Take(&recipient).Error
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("Unable to find recipient with smartID: %s", addressAndInfoRequest.RecipientSmartID.String))
			return
		}
	}

	senderAddress := &models.AddressAssignment{}
	if addressAndInfoRequest.SenderSmartID.Valid {
		senderAddress, err = senderAddress.FindPackageAddressWithSmartID(server.DB, sender, addressAndInfoRequest.Date.Time)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	}

	recipientAddress := &models.AddressAssignment{}
	if addressAndInfoRequest.RecipientSmartID.Valid {
		recipientAddress, err = recipientAddress.FindPackageAddressWithSmartID(server.DB, recipient, addressAndInfoRequest.Date.Time)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	}

	if addressAndInfoRequest.SenderSmartID.Valid && addressAndInfoRequest.RecipientSmartID.Valid {
		contacts := models.Contact{}
		err = contacts.SaveContacts(server.DB, sender.ID, recipient.ID)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	}

	packageDescription, err := addressAndInfoRequest.PackageDescription.SavePackageDescription(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	newPackage := models.Package{
		MailCarrierID:        reqUID,
		Tracking:             null.StringFrom(addressAndInfoRequest.Tracking),
		PackageDescriptionID: packageDescription.ID,
	}

	if addressAndInfoRequest.SenderSmartID.Valid {
		newPackage.SenderID = uuid.NullUUID{
			UUID:  sender.ID,
			Valid: true,
		}
	}

	if addressAndInfoRequest.RecipientSmartID.Valid {
		newPackage.RecipientID = uuid.NullUUID{
			UUID:  recipient.ID,
			Valid: true,
		}
	}

	err = newPackage.SavePackage(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	addressResponse := &responses.ToAndFromAddressSmartIDResponse{}

	if addressAndInfoRequest.SenderSmartID.Valid {
		senderResponse := responses.AddressSmartIDResponse{}
		responses.TranslateSmartAddressResponse(senderAddress, &senderResponse)
		addressResponse.Sender = senderResponse
	}

	if addressAndInfoRequest.RecipientSmartID.Valid {
		recipientResponse := responses.AddressSmartIDResponse{}
		responses.TranslateSmartAddressResponse(recipientAddress, &recipientResponse)
		addressResponse.Recipient = recipientResponse
	}

	responses.JSON(w, http.StatusCreated, addressResponse)
}

// GetPackageSenderAddressBySmartID retrieves a sender's package address using a customer's SmartID
func (server *Server) GetPackageSenderAddressBySmartID(w http.ResponseWriter, r *http.Request) {
	reqUID, permission, err := auth.ExtractAPIUserTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	reqUser := models.APIUser{}

	err = server.DB.Debug().Model(models.APIUser{}).Where("id = ?", reqUID).Take(&reqUser).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	hasPermission := models.FullAddressPermissions(reqUser.Permission)

	if !hasPermission || string(reqUser.Permission) != permission {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	vars := mux.Vars(r)
	smartID := sanitizeSmartID(strings.ToUpper(vars["smart_id"]))
	tracking := null.StringFromPtr(nil)
	trackingQuery, ok := vars["tracking"]
	if ok {
		tracking = null.StringFrom(trackingQuery)
	}
	date, err := time.Parse("2006-01-02", vars["date"])
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	user := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("smart_id = ?", smartID).Take(&user).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("Unable to find smartID: %s", smartID))
		return
	}

	address := models.AddressAssignment{}
	addressReceived, err := address.FindPackageAddressWithSmartID(server.DB, user, date)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	packageDescription := &models.PackageDescription{}
	packageDescription, err = packageDescription.SavePackageDescription(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	newPackage := models.Package{
		MailCarrierID: reqUID,
		SenderID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		RecipientID:          uuid.NullUUID{},
		Tracking:             tracking,
		PackageDescriptionID: packageDescription.ID,
	}
	err = newPackage.SavePackage(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	addressResponse := &responses.AddressSmartIDResponse{}
	responses.TranslateSmartAddressResponse(addressReceived, addressResponse)

	responses.JSON(w, http.StatusOK, addressResponse)
}

// GetPackageRecipientAddressBySmartID retrieves a recipient's package address using a customer's SmartID
func (server *Server) GetPackageRecipientAddressBySmartID(w http.ResponseWriter, r *http.Request) {
	reqUID, permission, err := auth.ExtractAPIUserTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	reqUser := models.APIUser{}

	err = server.DB.Debug().Model(models.APIUser{}).Where("id = ?", reqUID).Take(&reqUser).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	hasPermission := models.FullAddressPermissions(reqUser.Permission)

	if !hasPermission || string(reqUser.Permission) != permission {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	vars := mux.Vars(r)
	smartID := sanitizeSmartID(strings.ToUpper(vars["smart_id"]))
	tracking := null.StringFromPtr(nil)
	trackingQuery, ok := vars["tracking"]
	if ok {
		tracking = null.StringFrom(trackingQuery)
	}
	date, err := time.Parse("2006-01-02", vars["date"])
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	user := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("smart_id = ?", smartID).Take(&user).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("Unable to find smartID: %s", smartID))
		return
	}

	address := models.AddressAssignment{}
	addressReceived, err := address.FindPackageAddressWithSmartID(server.DB, user, date)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	packageDescription := &models.PackageDescription{}
	packageDescription, err = packageDescription.SavePackageDescription(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	newPackage := models.Package{
		MailCarrierID: reqUID,
		SenderID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		RecipientID:          uuid.NullUUID{},
		Tracking:             tracking,
		PackageDescriptionID: packageDescription.ID,
	}
	err = newPackage.SavePackage(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	addressResponse := &responses.AddressSmartIDResponse{}
	responses.TranslateSmartAddressResponse(addressReceived, addressResponse)

	responses.JSON(w, http.StatusOK, addressResponse)
}

// GetPackageZipBySmartID retrieves a user's zip code using a customer's SmartID
func (server *Server) GetPackageZipBySmartID(w http.ResponseWriter, r *http.Request) {
	reqUID, permission, err := auth.ExtractAPIUserTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	reqUser := models.APIUser{}

	err = server.DB.Debug().Model(models.APIUser{}).Where("id = ?", reqUID).Take(&reqUser).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	hasPermission := models.ZipPermissions(reqUser.Permission)

	if !hasPermission || string(reqUser.Permission) != permission {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	vars := mux.Vars(r)
	smartID := sanitizeSmartID(strings.ToUpper(vars["smart_id"]))
	date, err := time.Parse("2006-01-02", vars["date"])
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	user := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("smart_id = ?", smartID).Take(&user).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("Unable to find smartID: %s", smartID))
		return
	}

	address := models.AddressAssignment{}
	addressReceived, err := address.FindPackageAddressWithSmartID(server.DB, user, date)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	zipResponse := &responses.ZipResponse{}
	responses.TranslateZipResponse(addressReceived, zipResponse)

	responses.JSON(w, http.StatusOK, zipResponse)
}

// UpdateAddress updates the values of an address (this is used to fix errors with an address NOT change addresses)
func (server *Server) UpdateAddress(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	// Check if the address id is valid
	aaid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	//CHeck if the auth token is valid and  get the user id from it
	uid, err := auth.ExtractUITokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	// Check if the address assignment exist
	addressAssignment := models.AddressAssignment{}
	err = server.DB.Debug().Model(models.AddressAssignment{}).Where("id = ?", aaid).Take(&addressAssignment).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Address not found"))
		return
	}

	aid := addressAssignment.AddressID
	originalStart := addressAssignment.StartDate

	// If a user attempt to update a address not belonging to him
	if uid != addressAssignment.UserID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	// Read the data addressed
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	err = json.Unmarshal(body, &addressAssignment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	err = addressAssignment.Address.Update(server.DB, aid)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	err = addressAssignment.UpdateAddress(server.DB, aaid, originalStart)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	addresses, err := server.RetrieveAllAddresses(uid)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}

	response := responses.AddressesResponse{
		Addresses: addresses,
	}
	responses.JSON(w, http.StatusOK, response)
}

// DeleteAddress removes an address from the DB (not typically used)
func (server *Server) DeleteAddress(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	// Is a valid address id given to us?
	aid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Is this user authenticated?
	uid, err := auth.ExtractUITokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	// Check if the address exist
	addressAssignment := models.AddressAssignment{}
	err = server.DB.Debug().Model(models.Address{}).Where("id = ?", aid).Take(&addressAssignment).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Unauthorized"))
		return
	}

	// Is the authenticated user, the owner of this address?
	if uid != addressAssignment.UserID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	err = addressAssignment.DeleteAddress(server.DB, aid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	addresses, err := server.RetrieveAllAddresses(uid)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}

	response := responses.AddressesResponse{
		Addresses: addresses,
	}
	responses.JSON(w, http.StatusOK, response)
}

func geoLocate(addressAssignment *models.AddressAssignment) (err error) {
	geoCodeaddress := strings.ReplaceAll(addressAssignment.Address.LineOne+" "+addressAssignment.Address.City+" "+addressAssignment.Address.State+" "+addressAssignment.Address.ZipCode, " ", "+")

	res, err := http.Get("https://maps.googleapis.com/maps/api/geocode/json?address=" + geoCodeaddress + "&key=AIzaSyDCUjuA4aQIrKq8UQDaKnJPyc5cqxkzlPU")
	if err != nil {
		return
	}

	resBodyody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	var geoInfo geoInfo
	json.Unmarshal([]byte(resBodyody), &geoInfo)

	if len(geoInfo.Results) < 1 {
		err = fmt.Errorf("GeoLocation Error:  %s", geoInfo.ErrorMessage)
		return
	}

	addressAssignment.Address.Latitude = geoInfo.Results[0].Geometry.Location.Lat
	addressAssignment.Address.Longitude = geoInfo.Results[0].Geometry.Location.Lng

	return
}
