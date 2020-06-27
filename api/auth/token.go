package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
)

// CreateToken creates a login token that will be used by UI and API users
// expires after 1 hour
func CreateToken(userID uuid.UUID, authority string) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["permission"] = authority
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix() //Token expires after 1 hour
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("API_SECRET")))
}

// CreatePasswordResetToken creates a password reset token that will be included in a password reset link
// expires after 15 minutes
func CreatePasswordResetToken(userID uuid.UUID, password string) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["type"] = "reset"
	claims["user_id"] = userID
	claims["password"] = password
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix() //Token expires after 1 hour
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("API_SECRET")))
}

func TokenValid(r *http.Request) error {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		Pretty(claims)
	}
	return nil
}

func ExtractToken(r *http.Request) string {
	keys := r.URL.Query()
	token := keys.Get("token")
	if token != "" {
		return token
	}
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

func ExtractUITokenID(r *http.Request) (uuid.UUID, error) {

	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid && claims["permission"] == "ui" {
		uid, err := uuid.FromString(fmt.Sprintf("%s", claims["user_id"]))
		if err != nil {
			return uuid.UUID{}, err
		}
		return uid, nil
	}
	return uuid.UUID{}, nil
}

func ExtractAPIUserTokenID(r *http.Request) (uuid.UUID, string, error) {

	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return uuid.UUID{}, "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		uid, err := uuid.FromString(fmt.Sprintf("%s", claims["user_id"]))
		if err != nil {
			return uuid.UUID{}, "", err
		}
		return uid, fmt.Sprintf("%s", claims["permission"]), nil
	}
	return uuid.UUID{}, "", nil
}

func ExtractResetTokenID(r *http.Request) (uuid.UUID, string, error) {

	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return uuid.UUID{}, "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid && claims["type"] == "reset" {
		uid, err := uuid.FromString(fmt.Sprintf("%s", claims["user_id"]))
		if err != nil {
			return uuid.UUID{}, "", err
		}
		return uid, fmt.Sprintf("%s", claims["password"]), nil
	}
	return uuid.UUID{}, "", nil
}

//Pretty display the claims nicely in the terminal
func Pretty(data interface{}) {
	b, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(string(b))
}
