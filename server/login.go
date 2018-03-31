package main

import (
	"log"
	"net/http"
)

// login : login valida usuario, recibe usr (el usuario en claro) y pass (la contraseña en claro)
func login(usr string, pass string) bool {
	if usr == "Martin" && pass == "PASS" {
		return true
	}
	return false
}

// loginHandler : manejador de la peticion a /login
func loginHandler(w http.ResponseWriter, req *http.Request) {
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
