package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func chk(e error) {
	if e != nil {
		panic(e)
	}
}

// RunClient : run sincronice client
func RunClient() {

	/* creamos un cliente especial que no comprueba la validez de los certificados
	esto es necesario por que usamos certificados autofirmados (para pruebas) */
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// ** ejemplo de registro
	data := url.Values{}             // estructura para contener los valores
	data.Set("cmd", "hola")          // comando (string)
	data.Set("mensaje", "miusuario") // usuario (string)

	r, err := client.PostForm("https://localhost:8081", data) // enviamos por POST
	chk(err)
	io.Copy(os.Stdout, r.Body) // mostramos el cuerpo de la respuesta (es un reader)
	fmt.Println()
}
