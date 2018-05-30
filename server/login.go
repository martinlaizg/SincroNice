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
	user := types.User{}
	found := false
	for _, actUser := range users {
		if actUser.Email == email {
			found = true
			user = actUser
		}
	}
	if !found {
		log.Printf("Se ha intentado acceder con el usuario %s pero no existe\n", email)
		r.Status = false
		r.Msg = "Usuario y/o contraseña incorrectos"
		response(w, r)
		return
	}
	auth := crypto.ChkScrypt(user.Password, user.Salt, password)
	if !auth {
		log.Printf("Intento de login del usuario %s pero contraseña incorrecta", email)
		r.Status = false
		r.Msg = "Usuario y/o contraseña incorrectos"
		response(w, r)
		return
	}

	token := generateToken()
	sendToken(token, email)

	user.Token = token
	users[user.ID] = user

	resp := types.ResponseLogin{}
	resp.Status = true
	resp.Msg = "Logeado correctamente"
	resp.User = types.User{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}

	response(w, resp)
	log.Printf("Intento de acceso como %s, usuario y contraseña correcta, esperando verificación", email)
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")

	name := string(crypto.Decode64(req.Form.Get("name")))
	email := string(crypto.Decode64(req.Form.Get("email")))
	pass := crypto.Decode64(req.Form.Get("password"))

	dk, salt := crypto.Scrypt(pass)

	for _, user := range users {
		if user.Email == email {
			r.Status = false
			r.Msg = "Ya existe un usuario con el mismo nombre de usuario"
			log.Printf("Fail registry, user %v already exist", email)
			response(w, r)
			return
		}
	}
	userID := types.GenXid()
	folderID := types.GenXid()
	folder := types.Folder{
		ID:      folderID,
		UserID:  userID,
		Name:    "my-unit",
		Path:    "/",
		Created: time.Now().UTC().String(),
		Updated: time.Now().UTC().String(),
		Folders: make(map[string]string),
		Files:   make(map[string]string)}
	folders[folderID] = folder
	user := types.User{
		ID:         userID,
		Email:      email,
		Name:       name,
		Password:   dk,
		Salt:       salt,
		MainFolder: folderID}
	users[userID] = user
	r.Status = true
	r.Msg = "registrado correctamente"
	log.Printf("User %s registry successful", email)
	response(w, r)
}

func generateToken() string {
	token := make([]byte, 14)
	_, err := rand.Read(token)
	chk(err)
	sToken := crypto.Encode64(token)
	return sToken
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
	r := types.ResponseLogin{}
	w.Header().Set("Content-Type", "application/json")

	id := string(crypto.Decode64(req.Form.Get("id")))
	token := string(crypto.Decode64(req.Form.Get("token")))
	email := string(crypto.Decode64(req.Form.Get("email")))

	if chkToken(token, id) {
		log.Printf("Token del usuario %s verificado correctamente", email)
		user := users[id]
		r.ID = user.ID
		r.Email = user.Email
		r.Name = user.Name
		r.Token = user.Token
		r.MainFolder = user.MainFolder

		r.Status = true
		r.Msg = "Token correcto"

	} else {
		log.Printf("Token del usuario %s incorrecto", email)
		r.User = types.User{}
		r.Status = false
		r.Msg = "Token incorrecto"
	}
	response(w, r)
}

func chkToken(token string, id string) bool {
	return users[id].Token == token
}
