package modeltests

import (
	"log"
	"testing"

	_ "github.com/jinzhu/gorm/dialects/postgres" //postgres driver
	"github.com/nmelhado/pinpoint-api/api/models"
	"gopkg.in/go-playground/assert.v1"
)

func TestFindAllUsers(t *testing.T) {

	err := refreshTables()
	if err != nil {
		log.Fatalf("Error refreshing user table %v\n", err)
	}

	_, _, _, seedErr := seedTables()
	if seedErr != nil {
		log.Fatalf("Error seeding user table %v\n", err)
	}

	users, err := userInstance.FindAllUsers(server.DB)
	if err != nil {
		t.Errorf("this is the error getting the users: %v\n", err)
		return
	}
	assert.Equal(t, len(*users), 2)
}

func TestSaveUser(t *testing.T) {

	err := refreshTables()
	if err != nil {
		log.Fatalf("Error user refreshing table %v\n", err)
	}

	newUser := models.User{
		CosmoID:   "ABCDEFGH",
		FirstName: "Test",
		LastName:  "McGee",
		Phone:     "2125478965",
		Authority: models.UserAuth,
		Email:     "test@gmail.com",
		Password:  "password!",
	}
	savedUser, err := newUser.SaveUser(server.DB)
	if err != nil {
		t.Errorf("Error while saving a user: %v\n", err)
		return
	}
	assert.Equal(t, newUser.CosmoID, savedUser.CosmoID)
	assert.Equal(t, newUser.FirstName, savedUser.FirstName)
	assert.Equal(t, newUser.LastName, savedUser.LastName)
	assert.Equal(t, newUser.Phone, savedUser.Phone)
	assert.Equal(t, newUser.Authority, savedUser.Authority)
	assert.Equal(t, newUser.Email, savedUser.Email)
}

func TestGetUserByID(t *testing.T) {

	err := refreshTables()
	if err != nil {
		log.Fatalf("Error user refreshing table %v\n", err)
	}

	users, _, _, err := seedTables()
	if err != nil {
		log.Fatalf("cannot seed users table: %v", err)
	}
	foundUser, err := userInstance.FindUserByID(server.DB, users[0].ID)
	if err != nil {
		t.Errorf("this is the error getting one user: %v\n", err)
		return
	}
	assert.Equal(t, foundUser.ID, users[0].ID)
	assert.Equal(t, foundUser.CosmoID, users[0].CosmoID)
	assert.Equal(t, foundUser.FirstName, users[0].FirstName)
	assert.Equal(t, foundUser.LastName, users[0].LastName)
	assert.Equal(t, foundUser.Phone, users[0].Phone)
	assert.Equal(t, foundUser.Authority, users[0].Authority)
	assert.Equal(t, foundUser.Email, users[0].Email)
}

func TestUpdateAUser(t *testing.T) {

	err := refreshTables()
	if err != nil {
		log.Fatal(err)
	}

	users, _, _, err := seedTables()
	if err != nil {
		log.Fatalf("Cannot seed user: %v\n", err)
	}

	userUpdate := models.User{
		FirstName: "Al",
		Email:     "al@gmail.com",
		Password:  "NotTheJoker!",
	}
	updatedUser, err := userUpdate.UpdateAUser(server.DB, users[0].ID)
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.Equal(t, updatedUser.ID, users[0].ID)
	assert.Equal(t, updatedUser.CosmoID, users[0].CosmoID)
	assert.Equal(t, updatedUser.FirstName, userUpdate.FirstName)
	assert.Equal(t, updatedUser.LastName, users[0].LastName)
	assert.Equal(t, updatedUser.Phone, users[0].Phone)
	assert.Equal(t, updatedUser.Authority, users[0].Authority)
	assert.Equal(t, updatedUser.Email, userUpdate.Email)
}

func TestDeleteUser(t *testing.T) {

	err := refreshTables()
	if err != nil {
		log.Fatal(err)
	}

	users, _, _, err := seedTables()

	if err != nil {
		log.Fatalf("Cannot seed user: %v\n", err)
	}

	isDeleted, err := userInstance.DeleteUser(server.DB, users[0].ID)
	if err != nil {
		t.Errorf("this is the error deleting the user: %v\n", err)
		return
	}

	//one shows that the record has been deleted
	assert.Equal(t, isDeleted, int64(1))
}
