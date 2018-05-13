package main

import (
	"SincroNice/crypto"
	"SincroNice/types"
	"crypto/rand"
	"log"
	"net/http"
	"net/smtp"
	"time"
)

// loginHandler : manejador de la peticion a /login
func loginHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")

	email := string(crypto.Decode64(req.Form.Get("email")))
	password := crypto.Decode64(req.Form.Get("password"))
	user, exist := users[email]

	if !exist {
		r.Status = false
		r.Msg = "No existe ese usuario"
		log.Printf("Fail login, user %s not exist", email)
		response(w, r)
		return
	}
	auth := crypto.ChkScrypt(user.Password, user.Salt, password)

	if !auth {
		r.Status = false
		r.Msg = "Acceso denegado"
		response(w, r)
		log.Printf("Fail login, fail password for user %s", email)
		return
	}

	token := generateToken()
	sendToken(token, email)

	user.Token = token
	users[email] = user

	resp := types.ResponseToken{}
	resp.Status = true
	resp.Msg = "Logeado correctamente"
	resp.Token = token

	response(w, resp)
	log.Println("User " + email + " logging successful")
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")

	name := string(crypto.Decode64(req.Form.Get("name")))
	email := string(crypto.Decode64(req.Form.Get("email")))
	pass := crypto.Decode64(req.Form.Get("password"))

	dk, salt := crypto.Scrypt(pass)

	if _, exist := users[email]; exist {
		r.Status = false
		r.Msg = "Ya existe un usuario con el mismo nombre de usuario"
		log.Printf("Fail registry, user %v already exist", email)
		response(w, r)
		return
	}
	folder := types.Folder{
		UserID:  len(users) + 1,
		Name:    "my-unit",
		Path:    "/",
		Created: time.Now().UTC().String(),
		Updated: time.Now().UTC().String()}
	user := types.User{
		ID:         len(users) + 1,
		Name:       name,
		Password:   dk,
		Salt:       salt,
		MainFolder: &folder}
	users[email] = user
	r.Status = true
	r.Msg = "registrado correctamente"
	log.Printf("User %s registry successful", email)
	response(w, r)
}

func generateToken() string {
	token := make([]byte, 6)
	_, err := rand.Read(token)
	chk(err)
	sToken := crypto.Encode64(token)
	log.Println("Token generado ", sToken)
	return sToken
	// return "MyToken"
}

func sendToken(token string, to string) {

	from := "sincronicesl@gmail.com"
	pass := "Sincr0nice"
	subject := "Verificación de inicio de sesión"
	body := "Introduzca este código en su cliente: \"" + token + "\""

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
}

func checkTokenHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")

	email := string(crypto.Decode64(req.Form.Get("email")))
	token := string(crypto.Decode64(req.Form.Get("token")))

	if chkToken(token, email) {
		r.Status = true
		r.Msg = "Token correcto"
	} else {
		r.Status = false
		r.Msg = "Token incorrecto"
	}

	response(w, r)
}

func chkToken(token string, email string) bool {
	return users[email].Token == token
}
