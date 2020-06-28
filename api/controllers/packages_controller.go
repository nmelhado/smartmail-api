package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/nmelhado/smartmail-api/api/auth"
	"github.com/nmelhado/smartmail-api/api/models"
	"github.com/nmelhado/smartmail-api/api/responses"
	"github.com/nmelhado/smartmail-api/api/utils/formaterror"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
)

// GetPackages gets a user's packages
func (server *Server) GetPackages(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	reqID, err := uuid.FromString(vars["user_id"])
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, err := auth.ExtractUITokenID(r)
	if err != nil {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	if uid != reqID {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	packageDescriptionModel := models.PackageDescription{}

	openPackages, deliveredPackages, err := packageDescriptionModel.FindAndSortAllPackageDescriptionsForUser(server.DB, uid)

	packagesResponse := responses.TranslatePackagesResponse(*openPackages, *deliveredPackages)

	responses.JSON(w, http.StatusOK, packagesResponse)
}

// UpdatePackage is the struct for a package update request
type UpdatePackage struct {
	PackageID   uint64      `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	Tracking    string      `json:"tracking"`
	Description null.String `json:"description"`
	Delivered   bool        `json:"delivered"`
	DeliveredOn time.Time   `json:"delivered_on"`
}

// UpdatePackage gets a user's packages
func (server *Server) UpdatePackage(w http.ResponseWriter, r *http.Request) {
	uid, err := auth.ExtractUITokenID(r)
	if err != nil {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	packageToUpdate := UpdatePackage{}
	err = json.Unmarshal(body, &packageToUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	if uid != packageToUpdate.UserID {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	packageModel := models.Package{}

	updatedPackage, err := packageModel.UpdatePackage(server.DB, packageToUpdate.UserID, packageToUpdate.Tracking, packageToUpdate.Delivered, packageToUpdate.DeliveredOn)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}

	response := responses.UpdatePackageResponse{
		Success:   true,
		Tracking:  updatedPackage.Tracking.String,
		Delivered: updatedPackage.Delivered,
	}

	responses.JSON(w, http.StatusOK, response)
}

// UpdatePackageDescription gets a user's packages
func (server *Server) UpdatePackageDescription(w http.ResponseWriter, r *http.Request) {
	uid, err := auth.ExtractUITokenID(r)
	if err != nil {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	packageToUpdate := UpdatePackage{}
	err = json.Unmarshal(body, &packageToUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	if uid != packageToUpdate.UserID {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	packageDescriptionModel := models.PackageDescription{}

	updatedPackage, err := packageDescriptionModel.UpdatePackageDescription(server.DB, packageToUpdate.PackageID, packageToUpdate.Description)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}

	response := responses.UpdatePackageDescriptionResponse{
		Success:     true,
		ID:          updatedPackage.ID,
		Description: updatedPackage.Description,
	}

	responses.JSON(w, http.StatusOK, response)
}
