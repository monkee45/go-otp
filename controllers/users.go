package controllers

import (
	"log"
	"net/http"
	"net/url"

	"photos.com/models"
	"photos.com/utils"
	"photos.com/views"
)

type UsersController struct {
	NewUser    *views.View
	UserSignIn *views.View
	VerifyP    *views.View
	UserDBS    *models.UserDBService
}

type SignupForm struct {
	Name  string `schema:"name"`
	Email string `schema:"email"`
}

func NewUserController(newUserDBS models.UserDBService) *UsersController {
	return &UsersController{
		NewUser:    views.NewView("users/new.html"),
		UserSignIn: views.NewView("users/signin.html"),
		VerifyP:    views.NewView("users/verifyotp.html"),
		UserDBS:    &newUserDBS,
	}
}

func (u *UsersController) New(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	u.NewUser.Render(w, r, form)
}

func (u *UsersController) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	user := models.User{
		Name:  r.FormValue("name"),
		Email: r.FormValue("email"),
	}
	err := u.UserDBS.Create(&user)
	if err != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: err.Error(),
		}
		u.NewUser.Render(w, r, vd)
		return
	}
	vd.Alert = &views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Sign up successful!",
	}
	u.NewUser.Render(w, r, vd)
}

func (u *UsersController) SignIn(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	u.UserSignIn.Render(w, r, form)
}

func (u *UsersController) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	email := r.FormValue("email")
	user, err := u.UserDBS.FindByEmail(email)
	if err != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: err.Error(),
		}
		u.UserSignIn.Render(w, r, vd)
		return
	}
	otp := utils.GenerateRandomOTP()
	err = u.UserDBS.UpdateOTP(user.Email, otp)
	log.Printf("Generated OTP: %v\n", otp)
	if err != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: err.Error(),
		}
		u.UserSignIn.Render(w, r, vd)
		return
	}
	// send the otp to the users email address
	log.Printf("Redirecting...\n")
	// redirect to the the OTP validation screen/page
	http.Redirect(w, r, "/verify?email="+url.QueryEscape(user.Email), http.StatusSeeOther)
}

func (u *UsersController) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	email := r.URL.Query().Get("email")
	vd.User = &models.User{
		Email: email,
	}
	u.VerifyP.Render(w, r, vd)
}

func (u *UsersController) ProcessOTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("--> ProcessOTP\n")
	var vd views.Data
	email := r.FormValue("email")
	log.Printf("email from the form: %v\n", email)
	otp := r.FormValue("otp")

	log.Printf("about to call FindByEmail...\n")
	user, err := u.UserDBS.FindByEmail(email)
	if err != nil {
		log.Printf("FindByEmail error was not nil: %v\n", err.Error())
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: err.Error(),
		}
		u.VerifyP.Render(w, r, vd)
		return
	}
	if verifyOTP(user.Otp, otp) {
		otp = ""
		err = u.UserDBS.UpdateOTP(user.Email, otp)
		if err != nil {
			vd.Alert = &views.Alert{
				Level:   views.AlertLvlError,
				Message: err.Error(),
			}
			u.VerifyP.Render(w, r, vd)
			return
		}
	} else {
		log.Printf("invalid OTP entered\n")
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: "You entered an invalid One-Time Password",
		}
		vd.User = user
		u.VerifyP.Render(w, r, vd)
		return
	}
	http.Redirect(w, r, "/home", http.StatusFound)
}

//	func verifyOTP(hash []byte, otp string) bool {
//		return bcrypt.CompareHashAndPassword(hash, []byte(otp)) == nil
//	}
func verifyOTP(storedOTP string, otp string) bool {
	return storedOTP == otp
}
