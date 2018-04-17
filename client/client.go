package main

import (
	"SincroNice/crypto"
	"SincroNice/types"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/howeyc/gopass"
)

var baseURL = "https://localhost:8081"

var client *http.Client

func chk(e error) {
	if e != nil {
		panic(e)
	}
}

// send : enva {data} a la url localhost:8081/{endpoint}
func send(endpoint string, data url.Values) *http.Response {
	r, err := client.PostForm(baseURL+endpoint, data)
	chk(err)
	return r
}

func login() {
	fmt.Printf("\nLogin\n")
	fmt.Print("Email: ")
	var email string
	fmt.Scanln(&email)
	fmt.Print("Password: ")
	bpass, err := gopass.GetPasswdMasked()
	chk(err)

	log.Println("Acceso como " + email + "...\n")

	pass := crypto.Hash(bpass)

	data := url.Values{}
	data.Set("email", crypto.Encode64([]byte(email)))
	data.Set("password", crypto.Encode64(pass[:]))

	response := send("/login", data)
	bData, err := ioutil.ReadAll(response.Body)
	chk(err)
	var rData types.Response
	err = json.Unmarshal(bData, &rData)
	chk(err)

	if rData.Status == true {
		fmt.Printf("Logeado correctamente\n")
		return
	}
	fmt.Printf("Error al loguear: %v\n", rData.Msg)
}

func registry() {
	fmt.Printf("\nRegistro\n")
	fmt.Print("Nombre: ")
	var name string
	fmt.Scanln(&name)
	fmt.Print("Email: ")
	var email string
	fmt.Scanln(&email)
	fmt.Print("Contraseña: ")
	bpass, err := gopass.GetPasswdMasked() // Obtengo la contraseña
	chk(err)

	log.Println("Registrandose como " + email + "...\n")

	pass := crypto.Hash(bpass) // Hasheamos la contraseña con SHA512

	data := url.Values{}
	data.Set("name", crypto.Encode64([]byte(name)))
	data.Set("email", crypto.Encode64([]byte(email)))
	data.Set("password", crypto.Encode64(pass[:])) // Codificamos la contraseña en base64 para enviarla

	response := send("/register", data)
	bData, err := ioutil.ReadAll(response.Body)
	chk(err)
	var rData types.Response
	err = json.Unmarshal(bData, &rData)
	chk(err)

	if rData.Status == true {
		fmt.Printf("Registrado correctamente\n")
		return
	}
	fmt.Printf("Error al registrarse: %v\n", rData.Msg)
}

func menu() {

}

func createClient() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
}

// RunClient : run sincronice client
func main() {
	createClient()
	fmt.Printf("\nBienvenido a SincroNice\n\n")

	opt := ""
	for opt != "q" {

		fmt.Printf("1 - Login\n2 - Registro\nq - Salir\nOpcion: ")
		fmt.Scanf("%s\n", &opt)
		switch opt {
		case "1":
			login()
		case "2":
			registry()
		case "q":
			fmt.Println("Adios")
		default:
			fmt.Println("Intoduzca una opción correcta")
		}
		//menu()
	}
}
