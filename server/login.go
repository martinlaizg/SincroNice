package main

import (
	"SincroNice/crypto"
	"SincroNice/types"
	"encoding/json"
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
		resp, err := json.Marshal(user)
		chk(err)
		w.Write(resp)
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
		UserID:  len(users) + 1,
		Name:    "my-unit",
		Path:    "/",
		Created: time.Now().UTC().String(),
		Updated: time.Now().UTC().String()}
	folders[len(folders)+1] = folder
	user := types.User{
		ID:         len(users) + 1,
		Name:       name,
		Password:   dk,
		Salt:       salt,
		MainFolder: &folder}
	users[email] = user
	response(w, true, "registrado correctamente")
	log.Println("User " + email + " registry successful")
}
