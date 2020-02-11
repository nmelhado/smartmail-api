package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

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
	fmt.Printf("token id: %+v", uid)
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
	addressAssignment.AddressID = createAddress.ID

	addressAssignment.Prepare()
	err = addressAssignment.Validate()
	if err != nil {
		_, _ = addressAssignment.Address.DeleteAddress(server.DB, createAddress.ID)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	_, err = addressAssignment.SaveAddressAssignment(server.DB)
	if err != nil {
		_, _ = addressAssignment.Address.DeleteAddress(server.DB, createAddress.ID)
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, createAddress.ID))
	responses.JSON(w, http.StatusCreated, createAddress)
}

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

	_, err = addressAssignment.SaveAddressAssignment(server.DB)
	if err != nil {
		_, _ = addressAssignment.User.DeleteUser(server.DB, user.ID)
		_, _ = addressAssignment.Address.DeleteAddress(server.DB, createAddress.ID)
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, createAddress.ID))
	responses.JSON(w, http.StatusCreated, createAddress)
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

func (server *Server) GetMailingAddressByCosmoID(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	cosmoID, err := strconv.ParseUint(vars["cosmo_id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	date, err := time.Parse("2006-01-02", vars["date"])
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	user := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("cosmo_id = ?", cosmoID).Take(&user).Error

	reqUid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	reqUser := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("id = ?", reqUid).Take(&reqUser).Error

	if user.ID != reqUser.ID && reqUser.Authority != models.AdminAuth && reqUser.Authority != models.MailerAuth {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	address := models.AddressAssignment{}
	addressReceived, err := address.FindMailingAddressWithCosmo(server.DB, user, date)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, addressReceived)
}

func (server *Server) GetPackageAddressByCosmoID(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	cosmoID, err := strconv.ParseUint(vars["cosmo_id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	date, err := time.Parse("2006-01-02", vars["date"])
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	user := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("cosmo_id = ?", cosmoID).Take(&user).Error

	reqUid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	reqUser := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("id = ?", reqUid).Take(&reqUser).Error

	if user.ID != reqUser.ID && reqUser.Authority != models.AdminAuth && reqUser.Authority != models.MailerAuth {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	address := models.AddressAssignment{}
	addressReceived, err := address.FindPackageAddressWithCosmo(server.DB, user, date)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, addressReceived)
}

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

	// Check if the address assignment exist
	addressAssignment := models.AddressAssignment{}
	err = server.DB.Debug().Model(models.AddressAssignment{}).Where("address_id = ?", aid).Take(&addressAssignment).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Address not found"))
		return
	}

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

	// Start processing the request data
	addressUpdate := models.Address{}
	err = json.Unmarshal(body, &addressUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	addressUpdate.Prepare()
	err = addressUpdate.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	addressUpdate.ID = aid //this is important to tell the model the address id to update, the other update field are set above

	addressUpdated, err := addressUpdate.UpdateAddress(server.DB)

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
	addressAssignment := models.AddressAssignment{}
	err = server.DB.Debug().Model(models.Address{}).Where("address_id = ?", aid).Take(&addressAssignment).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Unauthorized"))
		return
	}

	// Is the authenticated user, the owner of this address?
	if uid != addressAssignment.UserID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	_, err = addressAssignment.Address.DeleteAddress(server.DB, aid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", aid))
	responses.JSON(w, http.StatusNoContent, "")
}
