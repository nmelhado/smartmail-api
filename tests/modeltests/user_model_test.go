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

	err = seedTables()
	if err != nil {
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
	models.User{
			ID: 1,
			CosmoID: "ABCDEFGH",
			FirstName: "Test",
			LastName: "McGee",
			Phone: "2125478965",
			Authority: models.UserAuth,
			Email:    "test@gmail.com",
			Password: "password!",
		},
	savedUser, err := newUser.SaveUser(server.DB)
	if err != nil {
		t.Errorf("Error while saving a user: %v\n", err)
		return
	}
	assert.Equal(t, newUser.ID, savedUser.ID)
	assert.Equal(t, newUser.Email, savedUser.Email)
	assert.Equal(t, newUser.Nickname, savedUser.Nickname)
}

func TestGetUserByID(t *testing.T) {

	err := refreshTables()
	if err != nil {
		log.Fatalf("Error user refreshing table %v\n", err)
	}

	user, err := seedTables()
	if err != nil {
		log.Fatalf("cannot seed users table: %v", err)
	}
	foundUser, err := userInstance.FindUserByID(server.DB, user.ID)
	if err != nil {
		t.Errorf("this is the error getting one user: %v\n", err)
		return
	}
	assert.Equal(t, foundUser.ID, user.ID)
	assert.Equal(t, foundUser.Email, user.Email)
	assert.Equal(t, foundUser.Nickname, user.Nickname)
}

func TestUpdateAUser(t *testing.T) {

	err := refreshTables()
	if err != nil {
		log.Fatal(err)
	}

	user, err := seedTables()
	if err != nil {
		log.Fatalf("Cannot seed user: %v\n", err)
	}

	userUpdate := models.User{
		ID:       1,
		Nickname: "modiUpdate",
		Email:    "modiupdate@gmail.com",
		Password: "password",
	}
	updatedUser, err := userUpdate.UpdateAUser(server.DB, user.ID)
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.Equal(t, updatedUser.ID, userUpdate.ID)
	assert.Equal(t, updatedUser.Email, userUpdate.Email)
	assert.Equal(t, updatedUser.Nickname, userUpdate.Nickname)
}

func TestDeleteUser(t *testing.T) {

	err := refreshTables()
	if err != nil {
		log.Fatal(err)
	}

	user, err := seedTables()

	if err != nil {
		log.Fatalf("Cannot seed user: %v\n", err)
	}

	isDeleted, err := userInstance.DeleteUser(server.DB, user.ID)
	if err != nil {
		t.Errorf("this is the error deleting the user: %v\n", err)
		return
	}
	//one shows that the record has been deleted or:
	// assert.Equal(t, int(isDeleted), 1)

	//Can be done this way too
	assert.Equal(t, isDeleted, int64(1))
}
