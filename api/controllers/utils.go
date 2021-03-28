package controllers

import (
	"errors"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/nmelhado/smartmail-api/api/models"

	uuid "github.com/satori/go.uuid"
)

var (
	UPS_INPUT = "UPS"
	UPS_USER = "UPS"
	USPS_INPUT = "USPS"
	USPS_USER = "USPS"
	FEDEX_INPUT = "FEDEX"
	FEDEX_USER = "FEDEX"
	LASERSHIP_INPUT = "LASERSHIP"
	LASERSHIP_USER = "LASERSHIP"
)

// Convert I to 1, S to 5, Z to 2, or O to 0)
func sanitizeSmartID(unsanitizedSmartID string) string {
	replacer := strings.NewReplacer("I", "1", "S", "5", "Z", "2", "O", "0")
	return replacer.Replace(unsanitizedSmartID)
}

func translateCarrier(db *gorm.DB, carrier string) (carrierID uuid.UUID, err error) {
	carrierUser  := models.APIUser{}

	switch strings.ToUpper(carrier) {
		case UPS_INPUT:
			carrierID, err = carrierUser.FindAPIUserIDByUsername(db, UPS_USER)
		case FEDEX_INPUT:
			carrierID, err = carrierUser.FindAPIUserIDByUsername(db, FEDEX_USER)
		case USPS_INPUT:
			carrierID, err = carrierUser.FindAPIUserIDByUsername(db, USPS_USER)
		case LASERSHIP_INPUT:
			carrierID, err = carrierUser.FindAPIUserIDByUsername(db, LASERSHIP_USER)
		default:
			err = errors.New("Invalid carrier")
		}

	return
}
