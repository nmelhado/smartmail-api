package controllers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/nmelhado/smartmail-api/api/auth"
	"github.com/nmelhado/smartmail-api/api/models"
	"github.com/nmelhado/smartmail-api/api/responses"
	"github.com/nmelhado/smartmail-api/api/utils/formaterror"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// Login validates a user and then calls SignIn
func (server *Server) Login(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	user.Prepare()
	err = user.Validate("login")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	token, validUser, err := server.SignIn(user.Email, user.Password)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}

	addresses, finalUser, err := server.RetrieveAllUserAddresses(validUser)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}

	contacts, err := server.pullContacts(validUser.ID)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	response := responses.UserAndAddressResponse{
		User:      finalUser,
		Addresses: addresses,
		Contacts:  contacts,
		Token:     token,
		Expires:   time.Now().Add(time.Hour * 1),
	}
	responses.JSON(w, http.StatusOK, response)
}

// Token validates a user and then calls SignIn
func (server *Server) Token(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	user.Prepare()
	err = user.Validate("login")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	token, _, err := server.SignIn(user.Email, user.Password)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}

	response := responses.TokenResponse{
		Token:   token,
		Expires: time.Now().Add(time.Hour * 1),
	}
	responses.JSON(w, http.StatusOK, response)
}

// RequestResetPassword finds a user by email and then returns a password reset token
func (server *Server) RequestResetPassword(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	requestedUser := models.User{}
	err = json.Unmarshal(body, &requestedUser)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	user := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("email = ?", requestedUser.Email).Take(&user).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	token, err := auth.CreatePasswordResetToken(user.ID, user.Password)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	response := responses.PasswordResetRequest{
		Token: token,
		Email: user.Email,
		Name:  user.FirstName,
	}
	responses.JSON(w, http.StatusOK, response)
}

// ResetPassword uses a token to reset a user's password
func (server *Server) ResetPassword(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	reqUID, checkPassword, err := auth.ExtractResetTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	user := &models.User{}

	user, err = user.FindUserByID(server.DB, reqUID)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	if user.Password != checkPassword {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("Unauthorized: this reset token has already been used"))
		return
	}

	err = json.Unmarshal(body, user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	user.Prepare()
	err = user.Validate("update")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	user, err = user.UpdateAUser(server.DB, reqUID)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	response := responses.PasswordReset{
		Success: true,
		Name:    user.FirstName,
	}
	responses.JSON(w, http.StatusOK, response)
}

// SignIn retrieves a token that is used for API endpoints
func (server *Server) SignIn(email, password string) (string, models.User, error) {

	var err error

	user := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("email = ?", email).Take(&user).Error
	if err != nil {
		return "", models.User{}, errors.New("User Not Found")
	}
	err = models.VerifyPassword(user.Password, password)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return "", models.User{}, err
	}
	token, err := auth.CreateToken(user.ID)
	return token, user, err
}

// RetrieveAllAddresses retrieves all non deleted addresses for a user
func (server *Server) RetrieveAllAddresses(userID uuid.UUID) (finalAddresses []responses.BasicAddress, err error) {
	addressAssignment := models.AddressAssignment{}

	addresses, err := addressAssignment.FindAllActiveAddressesForUser(server.DB, userID)
	if err != nil {
		return
	}

	finalAddresses = responses.TranslateAddresses(addresses)

	return
}

// RetrieveAllUserAddresses retrieves all non deleted addresses for a user
func (server *Server) RetrieveAllUserAddresses(user models.User) (finalAddresses []responses.BasicAddress, finalUser responses.CreateUserResponse, err error) {
	addressAssignment := models.AddressAssignment{}

	addresses, err := addressAssignment.FindAllActiveAddressesForUser(server.DB, user.ID)
	if err != nil {
		return
	}

	finalUser = responses.CreateUserResponse{
		ID:        user.ID,
		SmartID:   user.SmartID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt,
	}

	finalAddresses = responses.TranslateAddresses(addresses)

	return
}
