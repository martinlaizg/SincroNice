package main

import (
	"fmt"
	"net/http"
)

func response(w io.Writer, ok bool, msg string) {
	r := resp{Ok: ok, Msg: msg}    // formateamos respuesta
	rJSON, err := json.Marshal(&r) // codificamos en JSON
	chk(err)                       // comprobamos error
	w.Write(rJSON)                 // escribimos el JSON resultante
}

// login : login valida usuario, recibe usr (el usuario en claro) y pass (la contraseña en claro)
func login(usr string, pass string) bool {
	if usr == "Martin" && pass == "PASS" {
		return true
	}
	return false
}

// loginHandler : manejador de la peticion a /login
func loginHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()                              // es necesario parsear el formulario

	fmt.Println(req.Form.Get("usr"))
	fmt.Println(req.Form.Get("pass"))

	w.Header().Set("Content-Type", "text/plain") // cabecera estándar
	response.
}
