package controllers

import (
	"net/http"

	"github.com/nmelhado/smartmail-api/api/responses"
)

// Home provides a link to an API wiki
func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, "Welcome To the smartmail API! For documentation on how to use it, please visit: http://www.nmelhado.com.")
}
