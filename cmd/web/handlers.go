package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/jumaniyozov/gosub/data"
)

func (app *Config) HomePage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.gohtml", nil)
}

func (app *Config) LoginPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.gohtml", nil)
}

func (app *Config) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	_ = app.Session.RenewToken(r.Context())

	if err := r.ParseForm(); err != nil {
		app.ErrorLog.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.Models.User.GetByEmail(email)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Invalid credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	validPassword, err := user.PasswordMatches(password)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Invalid credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if !validPassword {
		msg := Message{
			To:      email,
			Subject: "Failed log in attempt",
			Data:    "Invalid log in attempt",
		}
		app.sendEmail(msg)
		app.Session.Put(r.Context(), "error", "Invalid credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "userID", user.ID)
	app.Session.Put(r.Context(), "user", user)

	app.Session.Put(r.Context(), "flash", "Successful login")

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (app *Config) Logout(w http.ResponseWriter, r *http.Request) {
	_ = app.Session.Destroy(r.Context())
	_ = app.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *Config) RegisterPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "register.page.gohtml", nil)
}

func (app *Config) PostRegisterPage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
	}

	u := data.User{
		Email:     r.Form.Get("email"),
		FirstName: r.Form.Get("first-name"),
		LastName:  r.Form.Get("last-name"),
		Password:  r.Form.Get("password"),
		Active:    0,
		IsAdmin:   0,
	}

	_, err = u.Insert(u)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Unable to create user")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	url := fmt.Sprintf("http://localhost:8080/activate?email=%s", u.Email)
	signedURL := GenerateTokenFromString(url)
	app.InfoLog.Println(signedURL)

	msg := Message{
		To:       u.Email,
		Subject:  "Activate your account",
		Template: "confirmation-email",
		Data:     template.HTML(signedURL),
	}

	app.sendEmail(msg)
	app.Session.Put(r.Context(), "flash", "Confirmation email sent. Check your email")
	http.Redirect(w, r, "/login", http.StatusSeeOther)

}

func (app *Config) ActivateAccount(w http.ResponseWriter, r *http.Request) {
	url := r.RequestURI
	testURL := fmt.Sprintf("http://localhost:8080%s", url)
	okay := VerifyToken(testURL)

	if !okay {
		app.Session.Put(r.Context(), "error", "Invalid token")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	u, err := app.Models.User.GetByEmail(r.URL.Query().Get("email"))
	if err != nil {
		app.Session.Put(r.Context(), "error", "No user found")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	u.Active = 1

	err = u.Update()
	if err != nil {
		app.Session.Put(r.Context(), "error", "Unable to update user.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "flash", "Account activated. You can now log in.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (ap *Config) SubscribeToPlan(w http.ResponseWriter, r *http.Request) {

}

func (app *Config) ChooseSubscription(w http.ResponseWriter, r *http.Request) {
	if !app.Session.Exists(r.Context(), "userID") {
		app.Session.Put(r.Context(), "warning", "You must be logged in to access this page")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	plans, err := app.Models.Plan.GetAll()
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	dataMap := make(map[string]any)

	dataMap["plans"] = plans

	app.render(w, r, "plans.page.gohtml", &TemplateData{
		Data: dataMap,
	})
}
