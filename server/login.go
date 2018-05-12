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

			r.Status = false
			r.Msg = "Acceso denegado"
			response(w, r)
			log.Printf("Fail login, fail password for user %s", email)
			return
		}
	}
	r.Status = false
	r.Msg = "No existe ese usuario"
	log.Printf("Fail login, user %s not exist", email)
	response(w, r)
	return
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

	// Map de prueba
	// var carpetas map[string]string
	// carpetas = make(map[string]string)
	// carpetas["bbrgdqhciteecb8ootig"] = "Prueba"

	userID := types.GenXid()
	folderID := types.GenXid()
	folder := types.Folder{
		ID:      folderID,
		UserID:  userID,
		Name:    "my-unit",
		Path:    "/",
		Created: time.Now().UTC().String(),
		Updated: time.Now().UTC().String(),
		Folders: make(map[string]string)}
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
