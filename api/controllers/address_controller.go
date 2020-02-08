package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nmelhado/pinpoint-api/api/auth"
	"github.com/nmelhado/pinpoint-api/api/models"
	"github.com/nmelhado/pinpoint-api/api/responses"
	"github.com/nmelhado/pinpoint-api/api/utils/formaterror"
)

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
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	if uid != addressAssignment.UserID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	user := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("id = ?", uid).Take(&user).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	addressAssignment.User = user

	createAddress, err := addressAssignment.Address.SaveAddress(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	addressAssignment.Address = *createAddress

	addressAssignment.Prepare()
	err = addressAssignment.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	_, err = addressAssignment.SaveAddressAssignment(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, createAddress.ID))
	responses.JSON(w, http.StatusCreated, createAddress)
}

func (server *Server) GetAddresss(w http.ResponseWriter, r *http.Request) {

	address := models.Address{}

	addresss, err := address.FindAllAddresss(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, addresss)
}

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

// !!!!!!!TODO!!!!!!!! Need to fix below as well as add a find by cosmo ID function

func (server *Server) UpdateAddress(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	// Check if the address id is valid
	aid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	//CHeck if the auth token is valid and  get the user id from it
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	// Check if the address exist
	address := models.Address{}
	err = server.DB.Debug().Model(models.Address{}).Where("id = ?", aid).Take(&address).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Address not found"))
		return
	}

	// If a user attempt to update a address not belonging to him
	if uid != address.AuthorID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	// Read the data addressed
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// Start processing the request data
	addressUpdate := models.Address{}
	err = json.Unmarshal(body, &addressUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	//Also check if the request user id is equal to the one gotten from token
	if uid != addressUpdate.AuthorID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	addressUpdate.Prepare()
	err = addressUpdate.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	addressUpdate.ID = address.ID //this is important to tell the model the address id to update, the other update field are set above

	addressUpdated, err := addressUpdate.UpdateAAddress(server.DB)

	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	responses.JSON(w, http.StatusOK, addressUpdated)
}

func (server *Server) DeleteAddress(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	// Is a valid address id given to us?
	aid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Is this user authenticated?
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	// Check if the address exist
	address := models.Address{}
	err = server.DB.Debug().Model(models.Address{}).Where("id = ?", aid).Take(&address).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Unauthorized"))
		return
	}

	// Is the authenticated user, the owner of this address?
	if uid != address.AuthorID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	_, err = address.DeleteAAddress(server.DB, aid, uid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", aid))
	responses.JSON(w, http.StatusNoContent, "")
}
