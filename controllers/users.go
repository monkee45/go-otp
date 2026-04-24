package controllers

import (
	"fmt"
	"net/http"

	"photos.com/models"
	"photos.com/views"
)

type UsersController struct {
	NewView *views.View
	UserDBS *models.UserDBService
}

type SignupForm struct {
	Name  string `schema:"name"`
	Email string `schema:"email"`
}

func NewUserController(newUserDBS models.UserDBService) *UsersController {
	return &UsersController{
		NewView: views.NewView("users/new.html"),
		UserDBS: &newUserDBS,
	}
}

func (u *UsersController) New(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	u.NewView.Render(w, r, form)
}

func (u *UsersController) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	user := models.User{
		Name:  r.FormValue("name"),
		Email: r.FormValue("email"),
	}
	fmt.Println(user)
	err := u.UserDBS.Create(&user)
	if err != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: err.Error(),
		}
		u.NewView.Render(w, r, vd)
		return
	}
	vd.Alert = &views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Sign up successful!",
	}
	u.NewView.Render(w, r, vd)
}
