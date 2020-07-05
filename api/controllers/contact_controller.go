package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nmelhado/smartmail-api/api/auth"
	"github.com/nmelhado/smartmail-api/api/models"
	"github.com/nmelhado/smartmail-api/api/responses"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
)

type addContactStruct struct {
	UserID  uuid.UUID `json:"user_id"`
	Contact contact   `json:"contact"`
}

type contact struct {
	SmartID string `json:"smart_id"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
}

// GetContacts gets a user's contacts
func (server *Server) GetContacts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := uuid.FromStringOrNil(vars["id"])
	var limit int64
	limitQuery, ok := vars["limit"]
	if ok {
		limitConvert, err := strconv.Atoi(limitQuery)
		if err != nil {
			fmt.Printf("Error, invalid limit:  %s", limitQuery)
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		limit = int64(limitConvert)
	}
	var offset int64
	pageQuery, ok := vars["page"]
	if ok {
		pageConvert, err := strconv.Atoi(pageQuery)
		if err != nil {
			fmt.Printf("Error, invalid limit:  %s", pageQuery)
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		offset = (int64(pageConvert) - 1) * limit
	}
	var sort string
	sortQuery, ok := vars["sort"]
	if ok {
		sort = sortQuery
	}
	var search null.String
	searchQuery, ok := vars["search"]
	if ok && strings.TrimSpace(searchQuery) != "" {
		search = null.StringFrom(searchQuery)
	}

	tokenID, err := auth.ExtractUITokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	if tokenID != uid {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	totalContacts, contacts, err := server.pullContacts(uid, limit, offset, sort, search)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	finalContacts := responses.Contacts{Contacts: contacts, Count: totalContacts, Success: true}

	responses.JSON(w, http.StatusOK, finalContacts)
}

// AddContact adds a contact to a user's contact list
func (server *Server) AddContact(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	addContact := addContactStruct{}
	err = json.Unmarshal(body, &addContact)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	uid, err := auth.ExtractUITokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	if uid != addContact.UserID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	contact := models.User{}

	if addContact.Contact.Email != "" && addContact.Contact.Phone != "" {
		err = server.DB.Debug().Model(models.User{}).Where("smart_id = ? AND email = ? AND phone = ?", addContact.Contact.SmartID, addContact.Contact.Email, addContact.Contact.Phone).Take(&contact).Error
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("Unable to find contact"))
			return
		}
	} else if addContact.Contact.Email != "" {
		err = server.DB.Debug().Model(models.User{}).Where("smart_id = ? AND email = ?", addContact.Contact.SmartID, addContact.Contact.Email).Take(&contact).Error
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("Unable to find contact"))
			return
		}
	} else if addContact.Contact.Phone != "" {
		err = server.DB.Debug().Model(models.User{}).Where("smart_id = ? AND phone = ?", addContact.Contact.SmartID, addContact.Contact.Phone).Take(&contact).Error
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("Unable to find contact"))
			return
		}
	} else {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unable to find contact"))
		return
	}

	newContact := models.Contact{}
	newContact, err = newContact.SaveContact(server.DB, addContact.UserID, contact.ID)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	fmt.Printf("contact: %+v", newContact)

	contactResponse := responses.TranslateContact(newContact)

	type singleContactResponse struct {
		Contact responses.Contact `json:"contact"`
	}

	finalResponse := &singleContactResponse{contactResponse}

	responses.JSON(w, http.StatusCreated, finalResponse)
}

func (server *Server) pullContacts(uid uuid.UUID, limit int64, offset int64, sort string, search null.String) (totalContacts int64, contacts []responses.Contact, err error) {
	var rawContacts []models.Contact
	if search.Valid {
		totalContacts, rawContacts, err = models.SearchContacts(server.DB, uid, limit, offset, search.String)
	} else {
		totalContacts, rawContacts, err = models.GetContacts(server.DB, uid, limit, offset, sort)
	}
	if err != nil {
		return
	}
	contacts = responses.TranslateContacts(rawContacts)
	return
}
