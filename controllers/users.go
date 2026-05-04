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
	user, err := u.UserService.FindByEmail(email)
	if err != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: err.Error(),
		}
		u.UserSignIn.Render(w, r, vd)
		return
	}
	err = generateAndSendOTP(u, user)
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

func (u *UsersController) ProcessOTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("--> UserController: ProcessOTP\n")
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
		http.Redirect(w, r, "/userhome", http.StatusFound)
	case "Resend code":
		err = generateAndSendOTP(u, user)
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
		u.VerifyP.Render(w, r, vd)
		return
	default:
		log.Println("No valid button value ")
	}
}

func (u *UsersController) Logout(w http.ResponseWriter, r *http.Request) {
	log.Printf("--> UserController: Logout()...\n")
	var vd views.Data
	session, ok := readSession(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	user := models.User{
		Email: session.Email,
	}
	clearSession(w)
	vd.User = &user
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

func generateAndSendOTP(u *UsersController, user *models.User) error {
	emailSubject := "Verification Code"
	otp := utils.GenerateRandomOTP()
	hash, _ := utils.HashOTP(otp)

	err := u.UserService.UpdateOTP(user.Email, hash)
	if err != nil {
		return err
	}
	// using a go routine as smtp.SendMail is very slow
	// not checking the return code - probably should use an error channel
	go mail.SendMail(user.Email, emailSubject, otp)
	return nil
}
