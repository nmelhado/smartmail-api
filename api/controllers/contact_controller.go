package controllers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/nmelhado/smartmail-api/api/auth"
	"github.com/nmelhado/smartmail-api/api/models"
	"github.com/nmelhado/smartmail-api/api/responses"
	uuid "github.com/satori/go.uuid"
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

	uid, err := auth.ExtractTokenID(r)
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

	contactResponse := responses.TranslateContact(newContact)

	type singleContactResponse struct {
		Contact responses.Contact `json:"contact"`
	}

	finalResponse := &singleContactResponse{contactResponse}

	responses.JSON(w, http.StatusCreated, finalResponse)
}
