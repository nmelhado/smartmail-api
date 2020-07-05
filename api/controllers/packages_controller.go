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
	"github.com/nmelhado/smartmail-api/api/utils/formaterror"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
)

// PreviewPackages gets a collection of a user's recent packages
func (server *Server) PreviewPackages(w http.ResponseWriter, r *http.Request) {

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

	packageModel := models.Package{}

	openPackages, deliveredPackages, err := packageModel.FindAllPackagesForUser(server.DB, uid)

	packagesResponse := responses.TranslatePackagesResponse(*openPackages, *deliveredPackages)

	responses.JSON(w, http.StatusOK, packagesResponse)
}

// GetPackages gets a collection of a user's recent packages
func (server *Server) GetPackages(w http.ResponseWriter, r *http.Request) {
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
	var packageType string
	packageTypeQuery, ok := vars["type"]
	if ok {
		packageType = packageTypeQuery
	}
	var search null.String
	searchQuery, ok := vars["search"]
	if ok && strings.TrimSpace(searchQuery) != "" {
		search = null.StringFrom(searchQuery)
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

	packageModel := models.Package{}

	openPackages, deliveredPackages, err := packageModel.FindPackagesForUser(server.DB, uid, packageType, limit, offset, search)

	packagesResponse := responses.TranslatePackagesResponse(*openPackages, *deliveredPackages)

	responses.JSON(w, http.StatusOK, packagesResponse)
}

// UpdatePackage is the struct for a package update request
type UpdatePackage struct {
	UserID            uuid.UUID `json:"user_id"`
	Tracking          string    `json:"tracking"`
	Delivered         bool      `json:"delivered"`
	DeliveredOn       null.Time `json:"delivered_on"`
	EstimatedDelivery null.Time `json:"estimated_delivery"`
}

// UpdatePackage update's a package's delivery delivered_on and estimated_delivery fields
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

	updatedPackage, err := packageModel.UpdatePackage(server.DB, packageToUpdate.UserID, packageToUpdate.Tracking, packageToUpdate.Delivered, packageToUpdate.DeliveredOn, packageToUpdate.EstimatedDelivery)
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

// UpdatePackageDescription is the struct for a package update description request
type UpdatePackageDescription struct {
	Tracking   string      `json:"tracking"`
	Contents   null.String `json:"contents"`
	OrderLink  null.String `json:"order_link"`
	OrderImage null.String `json:"order_image"`
}

// UpdatePackageDescription gets a user's packages
func (server *Server) UpdatePackageDescription(w http.ResponseWriter, r *http.Request) {
	uid, permission, err := auth.ExtractAPIUserTokenID(r)
	if err != nil {
		fmt.Print("\nUnauthorized\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	// Get the API user. This allows to get us to get the linked  smartmail account for the next step
	apiUser := &models.APIUser{}
	apiUser, err = apiUser.FindAPIUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	hasPermission := models.RetailPermissions(apiUser.Permission)

	if !hasPermission || string(apiUser.Permission) != permission {
		fmt.Print("\nUnauthorized 2\n")
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	packageToUpdate := UpdatePackageDescription{}
	err = json.Unmarshal(body, &packageToUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// Use the tracking number and the smartmail ID (obtained above) to get the package
	existingPackage := &models.Package{}
	existingPackage, err = existingPackage.FindPackageByTrackingAndShipper(server.DB, apiUser.SmartmailUser.ID, packageToUpdate.Tracking)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	updatedPackageDescription := &models.PackageDescription{}

	// Once all existing packages have been updated, remove this logic
	if existingPackage.PackageDescriptionID > 0 {
		updatedPackageDescription, err = updatedPackageDescription.UpdatePackageDescription(server.DB, existingPackage.PackageDescriptionID, packageToUpdate.Contents, packageToUpdate.OrderLink, packageToUpdate.OrderImage)
		if err != nil {
			formattedError := formaterror.FormatError(err.Error())
			responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
			return
		}
	} else {
		createPackageDescription := &models.PackageDescription{
			Contents:   packageToUpdate.Contents,
			OrderImage: packageToUpdate.OrderImage,
			OrderLink:  packageToUpdate.OrderLink,
		}
		updatedPackageDescription, err = createPackageDescription.SavePackageDescription(server.DB)
		if err != nil {
			formattedError := formaterror.FormatError(err.Error())
			responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
			return
		}
		// Assign a package description ID
		err = existingPackage.SetPackageDescription(server.DB, existingPackage.ID, updatedPackageDescription.ID)
		if err != nil {
			formattedError := formaterror.FormatError(err.Error())
			responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
			return
		}
	}

	response := responses.UpdatePackageDescriptionResponse{
		Success:    true,
		Contents:   updatedPackageDescription.Contents,
		OrderImage: updatedPackageDescription.OrderImage,
		OrderLink:  updatedPackageDescription.OrderLink,
	}

	responses.JSON(w, http.StatusOK, response)
}
