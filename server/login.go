package main

import (
	"SincroNice/crypto"
	"SincroNice/types"
	"log"
	"net/http"
	"time"
)

// loginHandler : manejador de la peticion a /login
func loginHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	w.Header().Set("Content-Type", "application/json")

	email := string(crypto.Decode64(req.Form.Get("email")))
	log.Println("Try login as " + email)
	password := crypto.Decode64(req.Form.Get("password"))
	user, exist := users[email]
	if !exist {
		response(w, false, "No existe ese usuario")
		log.Println("Fail login, user " + email + " not exist")
		return
	}
	auth := crypto.ChkScrypt(user.Password, user.Salt, password)

	if auth {
		response(w, true, "Acceso concedido")
		log.Println("User " + email + " logging successful")
		return
	}
	response(w, false, "Acceso denegado")
	log.Println("Fail login, fail password for user " + email)
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("User try registry")
	req.ParseForm()
	w.Header().Set("Content-Type", "application/json")

	name := string(crypto.Decode64(req.Form.Get("name")))
	email := string(crypto.Decode64(req.Form.Get("email")))
	pass := crypto.Decode64(req.Form.Get("password"))
	dk, salt := crypto.Scrypt(pass)

	if _, exist := users[email]; exist {
		response(w, false, "ya existe un usuario con el mismo nombre de usuario")
		log.Println("Fail registry, user " + email + " already exist")
		return
	}
	folder := types.Folder{
		ID:        len(folders) + 1,
		UserEmail: email,
		Name:      "my-unit",
		Path:      "/",
		Created:   time.Now().UTC().String(),
		Updated:   time.Now().UTC().String()}
	folders[email] = folder
	user := types.User{
		Name:       name,
		Password:   dk,
		Salt:       salt,
		MainFolder: &folder}
	users[email] = user
	response(w, true, "registrado correctamente")
	log.Println("User " + email + " registry successful")
}
