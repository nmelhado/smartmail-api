package controllers

import "github.com/nmelhado/smartmail-api/api/middlewares"

func (s *Server) initializeRoutes() {

	// Home Route
	s.Router.HandleFunc("/", middlewares.SetMiddlewareJSON(s.Home)).Methods("GET")

	// Login Route
	s.Router.HandleFunc("/login", middlewares.SetMiddlewareJSON(s.Login)).Methods("POST")

	// Forgot Password Route
	s.Router.HandleFunc("/forgot_password", middlewares.SetMiddlewareJSON(s.RequestResetPassword)).Methods("POST")
	s.Router.HandleFunc("/reset_password", middlewares.SetMiddlewareJSON(s.ResetPassword)).Methods("POST")

	// Token Route
	s.Router.HandleFunc("/token", middlewares.SetMiddlewareJSON(s.Token)).Methods("POST")

	//Sign up route
	s.Router.HandleFunc("/signup", middlewares.SetMiddlewareJSON(s.CreateUserAndAddress)).Methods("POST")

	//Contacts route
	s.Router.HandleFunc("/contacts/{id}", middlewares.SetMiddlewareJSON(s.GetContacts)).Methods("GET")
	s.Router.HandleFunc("/contacts", middlewares.SetMiddlewareJSON(s.AddContact)).Methods("POST")

	//Users routes
	s.Router.HandleFunc("/users", middlewares.SetMiddlewareJSON(s.CreateUser)).Methods("POST")
	s.Router.HandleFunc("/users", middlewares.SetMiddlewareJSON(s.GetUsers)).Methods("GET")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareJSON(s.GetUser)).Methods("GET")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.UpdateUser))).Methods("PUT")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareAuthentication(s.DeleteUser)).Methods("DELETE")

	// Addresses routes
	s.Router.HandleFunc("/addresses/mail/{sender_smart_id}/{recipient_smart_id}/{date}", middlewares.SetMiddlewareJSON(s.GetMailingAddressToAndFromBySmartID)).Methods("GET")
	s.Router.HandleFunc("/addresses/package/{sender_smart_id}/{recipient_smart_id}/{date}", middlewares.SetMiddlewareJSON(s.GetPackageAddressToAndFromBySmartID)).Methods("GET")

	// Zip routes
	s.Router.HandleFunc("/zip/mail/{smart_id}/{date}", middlewares.SetMiddlewareJSON(s.GetMailingZipBySmartID)).Methods("GET")
	s.Router.HandleFunc("/zip/package/{smart_id}/{date}", middlewares.SetMiddlewareJSON(s.GetPackageZipBySmartID)).Methods("GET")

	//Address routes
	s.Router.HandleFunc("/address", middlewares.SetMiddlewareJSON(s.CreateAddress)).Methods("POST")
	s.Router.HandleFunc("/address/{id}", middlewares.SetMiddlewareJSON(s.GetAddressByID)).Methods("GET")
	s.Router.HandleFunc("/address/mail/{smart_id}/{date}", middlewares.SetMiddlewareJSON(s.GetMailingAddressBySmartID)).Methods("GET")
	s.Router.HandleFunc("/address/package/{smart_id}/{date}", middlewares.SetMiddlewareJSON(s.GetPackageAddressBySmartID)).Methods("GET")
	s.Router.HandleFunc("/address/{id}", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.UpdateAddress))).Methods("PUT")
	s.Router.HandleFunc("/address/{id}", middlewares.SetMiddlewareAuthentication(s.DeleteAddress)).Methods("DELETE")
}
