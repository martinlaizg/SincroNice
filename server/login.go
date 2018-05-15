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
	for _, user := range users {
		if user.Email == email {
			auth := crypto.ChkScrypt(user.Password, user.Salt, password)

			if auth {
				resp, err := json.Marshal(user)
				chk(err)
				w.Write(resp)
				log.Println("User " + email + " logging successful")
				return
			}

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

	resp := types.ResponseLogin{}
	resp.Status = true
	resp.Msg = "Logeado correctamente"
	resp.Token = token
	resp.User = types.User{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}

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
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")

	email := string(crypto.Decode64(req.Form.Get("email")))
	token := string(crypto.Decode64(req.Form.Get("token")))

	if chkToken(token, email) {
		log.Println("Token verificado correctamente")
		r.Status = true
		r.Msg = "Token correcto"
	} else {
		log.Println("Token no verificado")
		r.Status = false
		r.Msg = "Token incorrecto"
	}
	response(w, r)
}

func chkToken(token string, email string) bool {
	return users[email].Token == token
}
