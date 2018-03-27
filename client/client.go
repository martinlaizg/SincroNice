package main

import (
	"fmt"

	"github.com/howeyc/gopass"
)

var baseURL = "https://localhost:8081"

func chk(e error) {
	if e != nil {
		panic(e)
	}
}

func menu() {
	fmt.Println("Bienvenido a SincroNice")
	fmt.Println("Login")
	fmt.Printf("Username: ")
	var usr string
	fmt.Scanf("%s\n", &usr)
	fmt.Printf("Pass: ")
	pass, err := gopass.GetPasswdMasked()
	chk(err)
	// var pass string
	// fmt.Scanf("%s\n", &pass)
	fmt.Println("Bienvenido " + usr)
	fmt.Println("Pass " + string(pass))
}

// RunClient : run sincronice client
func main() {
	menu()
	/* creamos un cliente especial que no comprueba la validez de los certificados
	esto es necesario por que usamos certificados autofirmados (para pruebas) */
	/* tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// ** ejemplo de registro
	data := url.Values{}      // estructura para contener los valores
	data.Set("usr", "Martin") // comando (string)
	data.Set("pass", "PASS")  // usuario (string)

	r, err := client.PostForm(baseURL+"/login", data) // enviamos por POST
	chk(err)
	io.Copy(os.Stdout, r.Body) // mostramos el cuerpo de la respuesta (es un reader)
	fmt.Println()*/
}
