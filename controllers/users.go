package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"photos.com/models"
	"photos.com/utils"
	"photos.com/views"
)

const sessionSecret = "super-secret-login-hash-key"

type UsersController struct {
	NewUser    *views.View
	UserSignIn *views.View
	UserHome   *views.View
	VerifyP    *views.View
	UserDBS    *models.UserDBService
}

type SignupForm struct {
	Name  string `schema:"name"`
	Email string `schema:"email"`
}

type SessionData struct {
	Email  string `json:"email"`
	Expiry int64  `json:"expiry"`
}

func NewUserController(newUserDBS models.UserDBService) *UsersController {
	return &UsersController{
		NewUser:    views.NewView("users/new.html"),
		UserSignIn: views.NewView("users/signin.html"),
		UserHome:   views.NewView("users/userhome.html"),
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

func (u *UsersController) Home(w http.ResponseWriter, r *http.Request) {
	u.UserHome.Render(w, r, nil)
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
	log.Printf("Ctlrs.ProcessSignIn: OTP: %v\n", otp)
	hash, _ := utils.HashOTP(otp)

	err = u.UserDBS.UpdateOTP(user.Email, hash)
	if err != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: err.Error(),
		}
		u.UserSignIn.Render(w, r, vd)
		return
	}
	// send the otp to the users email address
	cookie, _ := createSessionCookie(email)
	http.SetCookie(w, &cookie)

	// redirect to the the OTP validation screen/page
	http.Redirect(w, r, "/verify", http.StatusSeeOther)
}

func (u *UsersController) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("--> VerifyOTP()...\n")
	var vd views.Data
	session, ok := readSession(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	user := models.User{
		Email: session.Email,
	}
	vd.User = &user
	u.VerifyP.Render(w, r, vd)
}

func (u *UsersController) ProcessOTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("--> ProcessOTP\n")
	var vd views.Data
	email := r.FormValue("email")
	otp := r.FormValue("otp")

	user, err := u.UserDBS.FindByEmail(email)
	if err != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: err.Error(),
		}
		u.VerifyP.Render(w, r, vd)
		return
	}
	if verifyOTP(user.Otp, otp) {
		otp = ""
		err = u.UserDBS.UpdateOTP(user.Email, []byte(otp))
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
	http.Redirect(w, r, "/userhome", http.StatusFound)
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
