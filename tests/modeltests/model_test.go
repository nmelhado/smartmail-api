package modeltests

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"github.com/nmelhado/pinpoint-api/api/controllers"
	"github.com/nmelhado/pinpoint-api/api/models"
)

var server = controllers.Server{}
var userInstance = models.User{}
var postInstance = models.Post{}

func TestMain(m *testing.M) {
	var err error
	err = godotenv.Load(os.ExpandEnv("../../.env"))
	if err != nil {
		log.Fatalf("Error getting env %v\n", err)
	}
	Database()

	log.Printf("Before calling m.Run() !!!")
	ret := m.Run()
	log.Printf("After calling m.Run() !!!")
	//os.Exit(m.Run())
	os.Exit(ret)
}

func Database() {

	var err error

	TestDbDriver := os.Getenv("TestDbDriver")

		DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", os.Getenv("TestDbHost"), os.Getenv("TestDbPort"), os.Getenv("TestDbUser"), os.Getenv("TestDbName"), os.Getenv("TestDbPassword"))
		server.DB, err = gorm.Open(TestDbDriver, DBURL)
		if err != nil {
			fmt.Printf("Cannot connect to %s database\n", TestDbDriver)
			log.Fatal("This is the error:", err)
		} else {
			fmt.Printf("We are connected to the %s database\n", TestDbDriver)
		}

}

func refreshTables() error {
	err := server.DB.DropTableIfExists(&models.AddressAssignment{}, &models.User{},&models.Address{}).Error
	if err != nil {
		return err
	}
	err = server.DB.AutoMigrate(&models.User{}, &models.Address{},&models.AddressAssignment{}).Error
	if err != nil {
		return err
	}

	log.Printf("Successfully refreshed tables")
	return nil
}

func seedUsersAndAddresses() ([]models.User, []models.Address, []models.AddressAssignment, error) {

	var err error

	if err != nil {
		return []models.User, []models.Address, []models.AddressAssignment, err
	}
	
	var users = []models.User{
		models.User{
			CosmoID: "ABCDEFGH",
			FirstName: "Alfred",
			LastName: "Pennyworth",
			Phone: "2125478965",
			Authority: models.UserAuth,
			Email:    "alfred@gmail.com",
			Password: "BigScAryBats!",
		},
		models.User{
			CosmoID: "1B3D5F7H",
			FirstName: "Doug",
			LastName: "Mailman",
			Phone: "18005369377",
			Authority: models.MailerAuth,
			Email:    "Doug@pigeon.com",
			Password: "IL0veMai1!",
		},
		models.User{
			CosmoID: "12LIGHT3",
			FirstName: "Nikola",
			LastName: "Tesla",
			Phone: "3475389639",
			Authority: models.AdminAuth,
			Email:    "nikola@tesla.com",
			Password: "fuEdison!",
		},
	}

	var addresses = []models.Address{
		models.Address{
			Nickname:   "Work",
			LineOne: "347 Wayne Avenue",
			LineTwo: "Building B",
			UnitNumber: "52B",
			BusinessName: "Wayne Enterprises",
			BusinessName: "Bruce Wayne",
			City: "Gotham",
			State: "NY",
			ZipCode: "106745",
			Country: "United States",
		},
		models.Address{
			Nickname:   "Bat Cave",
			LineOne: "1 Martha Boulevard",
			City: "Gotham",
			State: "NY",
			ZipCode: "106744",
			Country: "United States",
		},
		models.Address{
			LineOne: "353 Main Street",
			City: "Dallas",
			State: "TX",
			ZipCode: "34567",
			Country: "United States",
		},
		models.Address{
			Nickname:   "Lab",
			LineOne: "26 Electric Avenue",
			City: "New York",
			State: "NY",
			ZipCode: "10021",
			Country: "United States",
		},
	}
	var addressesAssignments = []models.AddressAssignment{
		models.AddressAssignment{
			Status: models.LongTerm,
			time.Parse("2006-01-02", "2020-01-01"),
		},
		models.AddressAssignment{
			Status: models.Temporary,
			time.Parse("2006-01-02", "2020-04-01"),
			time.Parse("2006-01-02", "2020-07-01"),
		},
		models.AddressAssignment{
			Status: models.LongTerm,
			time.Parse("2006-01-02", "2020-01-01"),
		},
		models.AddressAssignment{
			Status: models.LongTerm,
			StartDate: time.Parse("2006-01-02", "2020-01-01"),
		},
	}
	for i, _ := range addresses {
		if i < len(addresses) - 1 {
		err = server.DB.Model(&models.User{}).Create(&users[i]).Error
		if err != nil {
			log.Fatalf("cannot seed users table: %v", err)
		}
		}
		err = server.DB.Model(&models.Address{}).Create(&addresses[i]).Error
		if err != nil {
			log.Fatalf("cannot seed address table: %v", err)
		}
		
		addressesAssignments[i].AddressID = address[i].ID
		if i < 2 {
		addressesAssignments[i].UserID = users[i].ID
		}
		if i >= 2 {
		addressesAssignments[i + 1].UserID = users[i].ID
		}

		err = server.DB.Model(&models.AddressAssignment{}).Create(&addressesAssignments[i]).Error
		if err != nil {
			log.Fatalf("cannot seed posts table: %v", err)
		}
	}
	return users, addresses, addressesAssignments, nil
}
