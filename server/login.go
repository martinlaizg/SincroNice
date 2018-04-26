package main

import (
	"SincroNice/crypto"
	"SincroNice/types"
	"fmt"
	"log"
	"net/http"
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

	if auth {
		r.Status = true
		r.Msg = "Acceso concedido"
		log.Printf("User %s logging successful", email)
		response(w, r)
		return
	}
	r.Status = false
	r.Msg = "Acceso denegado"
	response(w, r)
	log.Printf("Fail login, fail password for user %s", email)
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")

	name := string(crypto.Decode64(req.Form.Get("name")))
	email := string(crypto.Decode64(req.Form.Get("email")))
	pass := crypto.Decode64(req.Form.Get("password"))

	dk, salt := crypto.Scrypt(pass)

	fmt.Println("", name, email, string(pass), "")

	if _, exist := users[email]; exist {
		r.Status = false
		r.Msg = "Ya existe un usuario con el mismo nombre de usuario"
		log.Printf("Fail registry, user %v already exist", email)
		response(w, r)
		return
	}
	user := types.User{
		Name:     name,
		Password: dk,
		Salt:     salt}

	users[email] = user
	r.Status = true
	r.Msg = "registrado correctamente"
	log.Printf("User %s registry successful", email)
	response(w, r)
}
