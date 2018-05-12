package main

import (
	"SincroNice/crypto"
	"SincroNice/types"
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/howeyc/gopass"
)

var baseURL = "https://localhost:8081"

var client *http.Client

var usuario types.User
var folder types.Folder

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

func login() bool {
	fmt.Println("\nLogin\n")
	fmt.Print("Email: ")
	var email string
	fmt.Scanln(&email)
	fmt.Print("Password: ")
	bpass, err := gopass.GetPasswdMasked()
	chk(err)

	fmt.Printf("\nAcceso como %s...\n", email)

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

	if rData.ID != "" {
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

	fmt.Printf("\nRegistrandose como %v... \n", email)
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
	fmt.Printf("Error al registrarse: %v\n\n", rData.Msg)
}

func createClient() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
}

func explorarMiUnidad() bool {
	data := url.Values{}
	data.Set("id", crypto.Encode64([]byte(usuario.ID)))

	response := send("/u/{usuario.ID}/my-unit", data)
	bData, err := ioutil.ReadAll(response.Body)
	chk(err)
	var rData types.Folder
	err = json.Unmarshal(bData, &rData)
	chk(err)

	if rData.ID != "" {
		fmt.Println("\nSe encuentra en su directorio personal\n")
		folder = rData
		return true
	}
	fmt.Printf("Error al recuperar la carpeta personal: %v\n\n", rData)
	return false
}

func exploreFolder(id string) bool {
	data := url.Values{}
	data.Set("id", crypto.Encode64([]byte(usuario.ID)))
	data.Set("folderId", crypto.Encode64([]byte(id)))

	response := send("/u/{usuario.ID}/folders/{id}", data)
	bData, err := ioutil.ReadAll(response.Body)
	chk(err)
	var rData types.Folder
	err = json.Unmarshal(bData, &rData)
	chk(err)

	if rData.Folders != nil {
		fmt.Println("\nSe encuentra en el directorio " + rData.Name + "\n")
		folder = rData
		return true
	} else {
		fmt.Println("\nEl directorio que desea explorar, está vacío\n")
		return false
	}
	fmt.Printf("Error al recuperar la carpeta: %v\n\n", rData)
	return false
}

func exploredUnit() {
	opt := ""
	i := 1
	match := false
	var foldersIds map[int][]string
	foldersIds = make(map[int][]string)

	for opt != "q" {
		match = false
		for key, value := range folder.Folders {
			fmt.Println(i, "- "+value+" ("+key+")")
			foldersIds[i] = []string{key, value}
			i = i + 1
		}
		fmt.Printf("q - Salir\nOpcion: ")
		fmt.Scanf("%s\n", &opt)

		if opt != "q" {
			iter, err := strconv.Atoi(opt)
			if err != nil {
				fmt.Println("\nDebes introducir un número de la lista o q, ha introducido " + opt + "\n")
			} else {
				for key, value := range foldersIds {
					if key == iter {
						i = 1
						match = true
						if exploreFolder(value[0]) {
							exploredUnit()
						}
					}
				}
				if !match {
					fmt.Println("\nLa opción introducida no existe, debe escoger de entre la lista\n")
					i = 1
				}
			}
		}
	}
}

func loggedMenu() {
	fmt.Printf("\nBienvenido a su espacio personal " + usuario.Name + "\n")

	opt := ""
	for opt != "q" {
		fmt.Printf("\n1 - Explorar mi espacio\nq - Salir\nOpcion: ")
		fmt.Scanf("%s\n", &opt)
		switch opt {
		case "1":
			if explorarMiUnidad() {
				exploredUnit()
			}
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
		fmt.Printf("1 - Login\n2 - Registro\nq - Salir\nOpcion: ")
		fmt.Scanf("%s\n", &opt)
		switch opt {
		case "1":
			if login() {
				loggedMenu()
			}
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
