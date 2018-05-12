package main

import (
	"SincroNice/crypto"
	"SincroNice/types"
	"bytes"
	"crypto/sha256"
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"

	"github.com/howeyc/gopass"
)

var baseURL = "https://localhost:8081"

var client *http.Client

var usuario types.User

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

func subir() {

	fmt.Printf("\nRuta\n")
	var ruta string
	fmt.Scanln(&ruta)

	fmt.Printf("\nNombre del archivo\n")
	var nombre string
	fmt.Scanln(&nombre)

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", nombre)
	if err != nil {
		fmt.Println("error writing to buffer")
	}

	// open file handle
	fh, err := os.Open(ruta)
	if err != nil {
		fmt.Println("error opening file")
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	response, err := client.Post(baseURL+"/upload", contentType, bodyBuf)

	bData, err := ioutil.ReadAll(response.Body)
	chk(err)
	var rData types.Response
	err = json.Unmarshal(bData, &rData)
	chk(err)

	if rData.Status == true {
		fmt.Printf("Subido correctamente\n")
		return
	}
	fmt.Printf("Error al subir el archivo: %v\n", rData.Msg)

}

func login() bool {
	fmt.Printf("\nLogin\n")
	fmt.Print("Email: ")
	var email string
	fmt.Scanln(&email)
	fmt.Print("Password: ")
	bpass, err := gopass.GetPasswdMasked()
	chk(err)

	fmt.Printf("Acceso como %s...\n", email)

	pass := crypto.Hash(bpass)

	data := url.Values{}
	data.Set("email", crypto.Encode64([]byte(email)))
	data.Set("password", crypto.Encode64(pass[:]))

	response := send("/login", data)
	bData, err := ioutil.ReadAll(response.Body)
	chk(err)
	var rData types.User
	err = json.Unmarshal(bData, &rData)
	chk(err)

	if rData.MainFolder != nil {
		fmt.Printf("Logeado correctamente\n")
		usuario = rData
		return true
	}
	fmt.Printf("Error al loguear: %v\n\n", rData)
	return false
}

func registry() {
	fmt.Printf("\nRegistro\n")
	fmt.Print("Nombre: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	name := scanner.Text()

	fmt.Print("Email: ")
	scanner.Scan()
	email := scanner.Text()
	fmt.Print("Contraseña: ")
	bpass, err := gopass.GetPasswdMasked() // Obtengo la contraseña
	chk(err)

	fmt.Printf("Registrandose como %v \n", email)
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
		fmt.Printf("Registrado correctamente\n\n")
		return
	}
	fmt.Println(rData)
	fmt.Printf("Error al registrarse: %v\n\n", rData.Msg)
}

func createClient() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
}

func explorarMiUnidad() {
	fmt.Println("\nEsta es tu carpeta principal.")

}

func loggedMenu() {
	fmt.Printf("\nBienvenido a su espacio personal " + usuario.Name + "\n\n")

	opt := ""
	for opt != "q" {
		fmt.Printf("1 - Explorar mi espacio\nq - Salir\nOpcion: ")
		fmt.Scanf("%s\n", &opt)
		switch opt {
		case "1":
			explorarMiUnidad()
		case "q":
			fmt.Println("\nHasta la próxima " + usuario.Name + "\n")
		default:
			fmt.Println("\nIntoduzca una opción correcta")
		}
	}
}

// RunClient : run sincronice client
func main() {
	createClient()
	fmt.Printf("\nBienvenido a SincroNice\n\n")

	opt := ""
	for opt != "q" {
		fmt.Printf("1 - Login\n2 - Registro\n3 - Subir archivo\nq - Salir\nOpcion: ")
		fmt.Scanf("%s\n", &opt)
		switch opt {
		case "1":
			if login() {
				loggedMenu()
			}
		case "2":
			registry()
		case "3":
			subir()
		case "q":
			fmt.Println("Adios")
		default:
			fmt.Println("Intoduzca una opción correcta")
		}
		//menu()
	}
}
