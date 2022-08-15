package main

import "net/http"

func (app *Config) HomePage(w http.ResponseWriter, r *http.Request) {
	app.renderer(w, r, "home.page.gohtml", nil)
}

