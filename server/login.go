package main

import (
	"SincroNice/types"
	"log"
	"net/http"
)

// loginHandler : manejador de la peticion a /login
func loginHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("User try login")
	req.ParseForm() // es necesario parsear el formulario
	email := req.Form.Get("email")
	password := req.Form.Get("password")
	logged := users[email].Password == password
	msg := "OK"
	if logged == false {
		msg = "Usuario y/o contraseña incorrectos"
	}
	w.Header().Set("Content-Type", "application/json")
	response(w, logged, msg)
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("User try registry")
	req.ParseForm()
	//bsalt := make([]byte, 30)  // Crear salt de tamaño X
	//_, err := rand.Read(bsalt) // Generar salt aleatorio
	//chk(err)
	//salt := base64Encode(bsalt)
	email := req.Form.Get("email")
	registred := false
	msg := "Ya existe un usuario con ese nombre de usuario"
	_, exist := users[email]
	if !exist {
		registred = true
		msg = "Usuario registrado correctamente"
		user := types.User{
			Name:     req.Form.Get("name"),
			Password: req.Form.Get("password")}
		users[email] = user
	}
	w.Header().Set("Content-Type", "application/json")
	response(w, registred, msg)
}
