package main

import (
	"SincroNice/crypto"
	"SincroNice/types"
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

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

func subir() {

	fmt.Printf("\nFichero:")
	var ruta string
	fmt.Scanln(&ruta)

	//parts := make(map[string]byte[])

	carpetas := strings.Split(ruta, "/")
	nombre := carpetas[len(carpetas)-1]

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
	fmt.Print("\nLogin\n")
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
	rData := types.ResponseLogin{}
	err = json.Unmarshal(bData, &rData)
	chk(err)

	if !rData.Status {
		fmt.Println(rData.Msg)
		return false
	}
	usuario = rData.User
	return solicitarToken()
}

func solicitarToken() bool {
	fmt.Println("Introduzca el token que le hemos enviado por correo electrónico")
	fmt.Print("Token: ")
	var token string
	fmt.Scanln(&token)
	data := url.Values{}

	data.Set("id", crypto.Encode64([]byte(usuario.ID)))
	data.Set("email", crypto.Encode64([]byte(usuario.Email)))
	data.Set("token", crypto.Encode64([]byte(token)))

	response := send("/checkToken", data)

	respByte, err := ioutil.ReadAll(response.Body)
	chk(err)
	resp := types.ResponseLogin{}
	err = json.Unmarshal(respByte, &resp)
	chk(err)

	if resp.Status == true {
		usuario = resp.User
		fmt.Println("Sesión verificada correctamente")
	} else {
		usuario = types.User{}
		fmt.Println("El token introducido no coincide")
		return false
	}
	return true
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
	respByte, err := ioutil.ReadAll(response.Body)
	chk(err)
	resp := types.Response{}
	err = json.Unmarshal(respByte, &resp)
	chk(err)

	if resp.Status == true {
		fmt.Printf("Registrado correctamente\n\n")
		return
	}
	fmt.Printf("Error al registrarse: %v\n\n", resp.Msg)
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
		fmt.Printf("\nSe encuentra en su directorio personal\n")
		folder = rData
		return true
	}
	fmt.Printf("Error al recuperar la carpeta personal: %v\n\n", rData)
	return false
}

func getFolder(id string) bool {
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
	}
	fmt.Printf("Error al recuperar la carpeta: %v\n\n", rData)
	return false
}

func exploredUnit(mainfolder string) {
	opt := ""
	i := 1
	match := false
	var foldersIds map[int][]string
	foldersIds = make(map[int][]string)
	folderID := mainfolder
	for opt != "q" {
		_ = getFolder(folderID)

		match = false
		for key, value := range folder.Folders {
			fmt.Println(i, "- "+value+" ("+key+")")
			foldersIds[i] = []string{key, value}
			i = i + 1
		}
		fmt.Printf("s - Subir fichero\nq - Salir\nOpcion: ")
		fmt.Scanf("%s\n", &opt)

		switch opt {
		case "s":
			uploadFile()
		case "q":
			iter, err := strconv.Atoi(opt)
			if err != nil {
				fmt.Println("\nDebes introducir un número de la lista o q, ha introducido " + opt + "\n")
			} else {
				for key, value := range foldersIds {
					if key == iter {
						i = 1
						match = true
						folderID = value[0]
					}
				}
				if !match {
					fmt.Printf("\nLa opción introducida no existe, debe escoger de entre la lista\n")
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
		fmt.Printf("\n1 - Explorar mi espacio\nl - Logout\nq - Salir\nOpcion: ")
		fmt.Scanf("%s\n", &opt)
		switch opt {
		case "1":
			exploredUnit(usuario.MainFolder)
		case "l":
			fmt.Println("Cerrando sesión...")
			usuario = types.User{}
			opt = "q"
		case "q":
			fmt.Println("\nHasta la próxima " + usuario.Name + "\n")
		default:
			fmt.Println("\nIntoduzca una opción correcta")
		}
	}
}

func uploadFile() bool {
	fmt.Printf("Indique el fichero: ")
	path := "/home/martinlaizg/Desktop/doc.bat"
	// fmt.Scanf("%s\n", &path)
	tokens := strings.Split(path, "/")
	fileName := tokens[len(tokens)-1]

	fmt.Println(fileName)

	file, err := os.Open(path)
	chk(err)
	defer file.Close()
	fileInfo, _ := file.Stat()
	var fileSize int64 = fileInfo.Size()
	const fileChunk = 1 * (1 << 20) // 1 MB
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	fmt.Printf("Splitting to %d pieces.\n", totalPartsNum)
	fileChuked := make(map[[64]byte][]byte)
	parts := [][64]byte{}

	newPath := "/home/martinlaizg/Desktop/documento.sh"
	chk(err)

	for i := uint64(0); i < totalPartsNum; i++ {
		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)

		file.Read(partBuffer)

		hash := crypto.Hash(partBuffer)
		fileChuked[hash] = partBuffer
		parts = append(parts, hash)
		// // write to disk
		// fileName := "somebigfile_" + strconv.FormatUint(i, 10)
		// _, err := os.Create(fileName)
		// chk(err)
		// // write/save buffer to disk
		// fmt.Println("Split to : ", fileName)
	}
	_, err = os.Create(newPath)
	for _, value := range parts {
		ioutil.WriteFile(newPath, fileChuked[value], os.ModeAppend)
	}

	return true
}

// RunClient : run sincronice client
func main() {
	loadData()
	defer saveData()
	createClient()
	fmt.Printf("\nBienvenido a SincroNice\n\n")

	logged := false

	for opt := ""; opt != "q"; {
		if usuario.Token != "" {
			logged = true
		}

		if logged {
			loggedMenu()
			logged = false
		}
		if !logged {
			fmt.Printf("1 - Login\n2 - Registro\nq - Salir\nOpcion: ")
			fmt.Scanf("%s\n", &opt)
		}

		switch opt {
		case "1":
			logged = login()
		case "2":
			registry()
		case "q":
			fmt.Println("Adios")
		default:
			fmt.Println("Intoduzca una opción correcta")
		}
	}
}

func saveData() {
	log.Println("Saving data to JSON...")
	raw, err := json.Marshal(usuario)
	chk(err)
	err = ioutil.WriteFile("./userData.json", raw, 0777)
	chk(err)
	log.Println("Data saved")
}

func loadData() {
	log.Println("Loading data from JSON...")
	raw, err := ioutil.ReadFile("./userData.json")
	chk(err)
	err = json.Unmarshal(raw, &usuario)
	chk(err)
	log.Println("Data loaded")
}
