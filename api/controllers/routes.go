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

	// Sign up route
	s.Router.HandleFunc("/signup", middlewares.SetMiddlewareJSON(s.CreateUserAndAddress)).Methods("POST")

	// Contacts route
	s.Router.HandleFunc("/contacts/{id}", middlewares.SetMiddlewareJSON(s.GetContacts)).Queries("limit", "{limit}", "page", "{page}", "sort", "{sort}", "search", "{search}").Methods("GET")
	s.Router.HandleFunc("/contacts/{id}", middlewares.SetMiddlewareJSON(s.GetContacts)).Queries("limit", "{limit}", "page", "{page}", "sort", "{sort}").Methods("GET")
	s.Router.HandleFunc("/contacts", middlewares.SetMiddlewareJSON(s.AddContact)).Methods("POST")

	// API Users routes
	s.Router.HandleFunc("/api_users", middlewares.SetMiddlewareJSON(s.CreateAPIUser)).Methods("POST")
	s.Router.HandleFunc("/api_users", middlewares.SetMiddlewareJSON(s.GetAPIUsers)).Methods("GET")
	s.Router.HandleFunc("/api_users/{id}", middlewares.SetMiddlewareJSON(s.GetAPIUser)).Methods("GET")
	s.Router.HandleFunc("/api_users/{id}", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.UpdateAPIUser))).Methods("PUT")
	s.Router.HandleFunc("/api_users/{id}", middlewares.SetMiddlewareAuthentication(s.DeleteAPIUser)).Methods("DELETE")

	// Users routes
	s.Router.HandleFunc("/users", middlewares.SetMiddlewareJSON(s.CreateUser)).Methods("POST")
	s.Router.HandleFunc("/users", middlewares.SetMiddlewareJSON(s.GetUsers)).Methods("GET")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareJSON(s.GetUser)).Methods("GET")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.UpdateUser))).Methods("PUT")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareAuthentication(s.DeleteUser)).Methods("DELETE")

	// Mailing addresses sender and recipient routes
	s.Router.HandleFunc("/addresses/mail/{sender_smart_id}/{recipient_smart_id}/{date}", middlewares.SetMiddlewareJSON(s.GetMailingAddressToAndFromBySmartID)).Methods("GET")

	// Package addresses sender and recipient routes
	s.Router.HandleFunc("/addresses/package/{sender_smart_id}/{recipient_smart_id}/{date}/{tracking}", middlewares.SetMiddlewareJSON(s.GetPackageAddressToAndFromBySmartID)).Methods("GET")
	s.Router.HandleFunc("/addresses/package/{sender_smart_id}/{recipient_smart_id}/{date}", middlewares.SetMiddlewareJSON(s.GetPackageAddressToAndFromBySmartID)).Methods("GET")

	// Zip routes
	s.Router.HandleFunc("/zip/mail/{smart_id}/{date}", middlewares.SetMiddlewareJSON(s.GetMailingZipBySmartID)).Methods("GET")
	s.Router.HandleFunc("/zip/package/{smart_id}/{date}", middlewares.SetMiddlewareJSON(s.GetPackageZipBySmartID)).Methods("GET")

	// Address routes
	s.Router.HandleFunc("/address", middlewares.SetMiddlewareJSON(s.CreateAddress)).Methods("POST")
	s.Router.HandleFunc("/address/{id}", middlewares.SetMiddlewareJSON(s.GetAddressByID)).Methods("GET")
	s.Router.HandleFunc("/address/mail/{smart_id}/{date}", middlewares.SetMiddlewareJSON(s.GetMailingAddressBySmartID)).Methods("GET")
	s.Router.HandleFunc("/address/package/sender/{smart_id}/{date}/{tracking}", middlewares.SetMiddlewareJSON(s.GetPackageSenderAddressBySmartID)).Methods("GET")
	s.Router.HandleFunc("/address/package/sender/{smart_id}/{date}", middlewares.SetMiddlewareJSON(s.GetPackageSenderAddressBySmartID)).Methods("GET")
	s.Router.HandleFunc("/address/package/recipient/{smart_id}/{date}/{tracking}", middlewares.SetMiddlewareJSON(s.GetPackageRecipientAddressBySmartID)).Methods("GET")
	s.Router.HandleFunc("/address/package/recipient/{smart_id}/{date}", middlewares.SetMiddlewareJSON(s.GetPackageRecipientAddressBySmartID)).Methods("GET")
	s.Router.HandleFunc("/address/{id}", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.UpdateAddress))).Methods("PUT")
	s.Router.HandleFunc("/address/{id}", middlewares.SetMiddlewareAuthentication(s.DeleteAddress)).Methods("DELETE")

	// Packages route
	s.Router.HandleFunc("/preview_packages/{user_id}", middlewares.SetMiddlewareJSON(s.PreviewPackages)).Methods("GET")
	s.Router.HandleFunc("/package", middlewares.SetMiddlewareJSON(s.UpdatePackage)).Methods("Put")
	s.Router.HandleFunc("/package/description", middlewares.SetMiddlewareJSON(s.UpdatePackageDescription)).Methods("Put")
}
