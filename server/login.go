package main

import (
	"SincroNice/crypto"
	"SincroNice/types"
	"log"
	"net/http"
)

// loginHandler : manejador de la peticion a /login
func loginHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")
	defer response(w, r)

	email := string(crypto.Decode64(req.Form.Get("email")))
	log.Println("Try login as " + email)
	password := crypto.Decode64(req.Form.Get("password"))
	user, exist := users[email]
	if !exist {
		r.Status = false
		r.Msg = "No existe ese usuario"
		log.Println("Fail login, user " + email + " not exist")
		return
	}
	auth := crypto.ChkScrypt(user.Password, user.Salt, password)

	if auth {
		r.Status = true
		r.Msg = "Acceso concedido"
		log.Println("User " + email + " logging successful")
		return
	}
	r.Status = false
	r.Msg = "Acceso denegado"
	log.Println("Fail login, fail password for user " + email)
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("User try registry")
	req.ParseForm()

	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")
	defer response(w, r)

	name := string(crypto.Decode64(req.Form.Get("name")))
	email := string(crypto.Decode64(req.Form.Get("email")))
	pass := crypto.Decode64(req.Form.Get("password"))
	dk, salt := crypto.Scrypt(pass)

	if _, exist := users[email]; exist {
		r.Status = false
		r.Msg = "Ya existe un usuario con el mismo nombre de usuario"
		log.Println("Fail registry, user " + email + " already exist")
		return
	}
	user := types.User{
		Name:     name,
		Password: dk,
		Salt:     salt}
	users[email] = user
	r.Status = true
	r.Msg = "registrado correctamente"
	log.Println("User " + email + " registry successful")
}
