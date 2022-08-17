package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(app.SessionLoad)

	mux.Get("/", app.HomePage)
	mux.Post("/login", app.LoginPage)
	mux.Get("/logout", app.Logout)
	mux.Get("/register", app.RegisterPage)
	mux.Post("/register", app.PostRegisterPage)
	mux.Get("/activate", app.ActivateAccount)
	mux.Get("/plans", app.ChooseSubscription)
	mux.Get("/subscribe", app.SubscribeToPlan)


	return mux
}
