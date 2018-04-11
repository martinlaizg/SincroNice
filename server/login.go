package main

import (
	"SincroNice/types"
	"crypto/rand"
	"log"
	"net/http"
)

// login : login valida usuario, recibe usr (el usuario en claro) y pass (la contraseña en claro)
func login(usr string, pass string) bool {
	for _, user := range users {
		if user.Password == pass {
			return true
		}
	}
	return false
}

// loginHandler : manejador de la peticion a /login
func loginHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("User try login")
	req.ParseForm() // es necesario parsear el formulario
	usr := req.Form.Get("usr")
	pass := req.Form.Get("pass")
	log.Println(usr, pass)
	logged := login(usr, pass)
	msg := "OK"
	if logged == false {
		msg = "Usuario y contraseña incorrectos"
	}
	w.Header().Set("Content-Type", "application/json")
	response(w, logged, msg)

}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("User try registry")
	req.ParseForm()
	bsalt := make([]byte, 30)  // Crear salt de tamaño X
	_, err := rand.Read(bsalt) // Generar salt aleatorio
	chk(err)
	salt := base64Encode(bsalt)

	usr := types.User{
		ID:       salt,
		Username: req.Form.Get("username"),
		Name:     req.Form.Get("name"),
		Password: req.Form.Get("password")}
	users[string(salt)] = usr
}
