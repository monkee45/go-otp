package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"photos.com/mail"
	"photos.com/models"
	"photos.com/utils"
	"photos.com/views"
)

const sessionSecret = "super-secret-login-hash-key"

type UsersController struct {
	NewUser     *views.View
	UserSignIn  *views.View
	UserHome    *views.View
	VerifyP     *views.View
	UserService *models.UserService
}

type SignupForm struct {
	Name  string `schema:"name"`
	Email string `schema:"email"`
}

type SessionData struct {
	Email  string `json:"email"`
	Expiry int64  `json:"expiry"`
}

func NewUserController(newUserService models.UserService) *UsersController {
	return &UsersController{
		NewUser:     views.NewView("users/new.html"),
		UserSignIn:  views.NewView("users/signin.html"),
		UserHome:    views.NewView("users/userhome.html"),
		VerifyP:     views.NewView("users/verifyotp.html"),
		UserService: &newUserService,
	}
}

// New is used to render the form where a user can create a new user account
//
// GET /signup
func (u *UsersController) New(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	u.NewUser.Render(w, r, form)
}

// ****** CREATE ******
// Create is used to process the signup form when a user submits it.
// This is used to create a new User account
//
// POST /signin
func (u *UsersController) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	user := models.User{
		Name:  r.FormValue("name"),
		Email: r.FormValue("email"),
	}
	err := u.UserService.Create(&user)
	if err != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: err.Error(),
		}
		vd.User = &models.User{
			Name:  user.Name,
			Email: user.Email,
		}
		u.NewUser.Render(w, r, vd)
		return
	}
	vd.Alert = &views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Sign up successful!",
	}

	u.UserSignIn.Render(w, r, vd)
}

// ***** SignIn *****
// SignIn is used to render the first stage of the login process
// GET /signin

func (u *UsersController) SignIn(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	u.UserSignIn.Render(w, r, form)
}

// ***** ProcessSignIn *****
// ProcessSignIn is used to validate the email, generate and
// send a one-time password to the users email, create the session cookie
// and redirect to the OTP entry screen
// POST /signin

func (u *UsersController) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	email := r.FormValue("email")
	user, err := u.UserService.FindByEmail(email)
	if err != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: err.Error(),
		}
		u.UserSignIn.Render(w, r, vd)
		return
	}
	err = generateAndSendOTP(u, user, u.UserService.Logger)
	if err != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: err.Error(),
		}
		u.UserSignIn.Render(w, r, vd)
		return
	}

	cookie, _ := createSessionCookie(email)
	http.SetCookie(w, &cookie)

	// redirect to the the OTP validation screen/page
	http.Redirect(w, r, "/verify", http.StatusSeeOther)
}

// ***** VerifyOTP *****
// VerifyOTP gets the user email from the session cookie
// It re-displays the email address in the email entry field and adds
// a new field for the entry of the One-Time Password
// GET /verify

func (u *UsersController) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	session, ok := readSession(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	user := models.User{
		Email: session.Email,
	}
	vd.User = &user
	vd.Alert = &views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Check your email for a verification code",
	}
	u.VerifyP.Render(w, r, vd)
}

// ***** ConfirmOTP *****
// ConfirmOTP is used to get the user-entered OTP value and determine
// if they have pressed the "SignIn" or "Resend Code" button
// In the case of "Resend", we sent a completely new OTP
// POST /verify

func (u *UsersController) ConfirmOTP(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	_ = r.ParseForm()
	email := r.FormValue("email")
	otp := r.FormValue("otp")
	buttonValue := r.FormValue("action")

	user, err := u.UserService.FindByEmail(email)
	if err != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: err.Error(),
		}
		vd.User = &models.User{
			Email: email,
		}
		u.VerifyP.Render(w, r, vd)
		return
	}

	switch buttonValue {
	case "Sign In":
		if time.Now().After(user.OtpExpiry) {
			vd.Alert = &views.Alert{
				Level:   views.AlertLvlError,
				Message: "OTP Code has Expired",
			}
			u.UserSignIn.Render(w, r, vd)
			return

		}
		if verifyOTP(user.Otp, otp) {
			otp = ""
			err = u.UserService.UpdateOTP(user.Email, []byte(otp))
			if err != nil {
				vd.Alert = &views.Alert{
					Level:   views.AlertLvlError,
					Message: err.Error(),
				}
				u.VerifyP.Render(w, r, vd)
				return
			}
		} else {
			vd.Alert = &views.Alert{
				Level:   views.AlertLvlError,
				Message: "You entered an invalid One-Time Password",
			}
			vd.User = user
			u.VerifyP.Render(w, r, vd)
			return
		}
		vd.User = user
		fmt.Println("Rendering User Home after successful login")
		http.Redirect(w, r, "/userhome", http.StatusFound)
		// u.UserHome.Render(w, r, vd)
	case "Resend code":
		err = generateAndSendOTP(u, user, u.UserService.Logger)
		if err != nil {
			vd.Alert = &views.Alert{
				Level:   views.AlertLvlError,
				Message: err.Error(),
			}
			u.UserSignIn.Render(w, r, vd)
			return
		}
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlSuccess,
			Message: "New One Time Password Sent",
		}
		vd.User = user
		fmt.Println("Re-Rendering VerifyOTP after new code request")
		u.VerifyP.Render(w, r, vd)
	default:
	}
}

func (u *UsersController) Home(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	session, ok := readSession(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	user, err := u.UserService.FindByEmail(session.Email)
	if err != nil {
		http.Error(w, "Something went wrong!", http.StatusInternalServerError)
		return
	}
	vd.User = user
	u.UserHome.Render(w, r, vd)
}

func (u *UsersController) Logout(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	session, ok := readSession(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	user := models.User{
		Email: session.Email,
	}
	clearSession(w)
	vd.User = nil
	vd.Alert = &views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: user.Email + "successfully logged out",
	}
	u.UserSignIn.Render(w, r, vd)
}

func RequireSession(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		session, ok := readSession(r)
		if !ok || session.Email == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	})
}

func verifyOTP(hash []byte, otp string) bool {
	return bcrypt.CompareHashAndPassword(hash, []byte(otp)) == nil
}

// func verifyOTP(storedOTP string, otp string) bool {
// 	return storedOTP == otp
// }

// Create the session cookie
func createSessionCookie(email string) (http.Cookie, error) {
	data := SessionData{
		Email: email,
		// gets the current time, adds 10 minutes to it, and then converts
		// that new time to a Unix timestamp, which is the number of
		// seconds since January 1, 1970
		Expiry: time.Now().Add(10 * time.Minute).Unix(),
	}
	jsonData, _ := json.Marshal(data)
	payload := base64.StdEncoding.EncodeToString(jsonData)

	mac := hmac.New(sha256.New, []byte(sessionSecret))
	mac.Write([]byte(payload))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	value := payload + "." + signature

	return http.Cookie{
		Name:     "session",
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

// Verify and read the cookie
func readSession(r *http.Request) (*SessionData, bool) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, false
	}

	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return nil, false
	}
	payload, sig := parts[0], parts[1]

	mac := hmac.New(sha256.New, []byte(sessionSecret))
	mac.Write([]byte(payload))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return nil, false
	}

	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, false
	}

	var data SessionData
	err = json.Unmarshal(decoded, &data)
	if err != nil {
		return nil, false
	}
	if time.Now().Unix() > data.Expiry {
		return nil, false
	}
	return &data, true
}

func clearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func generateAndSendOTP(u *UsersController, user *models.User, logger *slog.Logger) error {
	emailSubject := "Verification Code"
	otp := utils.GenerateRandomOTP()
	hash, _ := utils.HashOTP(otp)

	err := u.UserService.UpdateOTP(user.Email, hash)
	if err != nil {
		return err
	}
	fmt.Println("OTP Code: ", otp)
	// using a go routine as smtp.SendMail is very slow
	// not checking the return code - probably should use an error channel
	go mail.SendMail(user.Email, emailSubject, otp, logger)
	return nil
}
