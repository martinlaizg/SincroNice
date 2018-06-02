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

func subirDrive() bool {
	fmt.Printf("\nRuta\n")
	var ruta string
	fmt.Scanln(&ruta)
	//ruta = "C:/prueba.ppt"
	fmt.Printf("\nNombre del archivo\n")
	var nombre string
	fmt.Scanln(&nombre)
	///////
	file, err := os.Open(ruta)
	chk(err)
	defer file.Close()

	fileInfo, _ := file.Stat()
	var fileSize int64 = fileInfo.Size()
	const fileChunk = 1 * (1 << 20) // 1 MB
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	version := types.Version{
		ID: types.GenXid(),
	}

	for i := uint64(0); i < totalPartsNum; i++ {
		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)

		file.Read(partBuffer)

		blockID := checkBlock(partBuffer)
		version.Blocks = append(version.Blocks, blockID)
	}
	fileT := types.File{
		FolderID: folder.ID,
		Name:     nombre,
		OwnerID:  usuario.ID,
		Versions: append(make([]types.Version, 0), version),
	}

	return uploadFileT(fileT)
}

func uploadFileT(file types.File) bool {
	data := url.Values{}
	fileB, err := json.Marshal(file)
	chk(err)
	data.Set("file", crypto.Encode64(fileB))

	resp := send("/checkBlock", data)
	bData, err := ioutil.ReadAll(resp.Body)
	chk(err)
	response := types.Response{}
	err = json.Unmarshal(bData, &response)
	chk(err)

	return true
}

func checkBlock(buffer []byte) string {

	hash := crypto.Hash(buffer)
	hash64 := crypto.Encode64(hash[:])
	data := url.Values{}
	data.Set("hash", hash64)
	resp := send("/checkBlock", data)
	bData, err := ioutil.ReadAll(resp.Body)
	chk(err)
	response := types.Response{}
	err = json.Unmarshal(bData, &response)
	chk(err)
	blockID := response.Msg

	if !response.Status {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("fileupload", blockID)
		chk(err)
		_, err = io.Copy(part, bytes.NewReader(buffer))
		chk(err)
		_ = writer.WriteField("userID", crypto.Encode64([]byte("Luis")))
		_ = writer.WriteField("folderID", crypto.Encode64([]byte("")))
		_ = writer.WriteField("blockID", crypto.Encode64([]byte(blockID)))
		err = writer.Close()
		chk(err)
		req, err := http.NewRequest("POST", baseURL+"/uploadDrive", body)
		chk(err)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := client.Do(req)
		chk(err)
		bData, err := ioutil.ReadAll(resp.Body)
		chk(err)
		response := types.Response{}
		err = json.Unmarshal(bData, &response)
		chk(err)

	}
	return blockID
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
	loadData()
	defer saveData()
	createClient()
	fmt.Printf("\nBienvenido a SincroNice\n\n")

	opt := ""
	for opt != "q" {
		if usuario.Token != "" {
			loggedMenu()
			opt = "q"
		}
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
			subirDrive()
		case "4":
			subirDrive()
		case "q":
			fmt.Println("Adios")
		default:
			fmt.Println("Intoduzca una opción correcta")
		}
		//menu()
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
