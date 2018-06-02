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

	"github.com/fatih/color"
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

func getFolder(id string) bool {
	data := url.Values{}
	data.Set("id", crypto.Encode64([]byte(usuario.ID)))
	data.Set("token", crypto.Encode64([]byte(usuario.Token)))
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

func crearCarpeta(actualFolder string) bool {
	fmt.Print("Introduzca el nombre de la carpeta: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	folderName := scanner.Text()

	fmt.Printf("\nCreando la carpeta con el nombre %s...", folderName)
	data := url.Values{}
	data.Set("user", crypto.Encode64([]byte(usuario.ID)))
	data.Set("token", crypto.Encode64([]byte(usuario.Token)))
	data.Set("actualFolder", crypto.Encode64([]byte(actualFolder)))
	data.Set("folderName", crypto.Encode64([]byte(folderName)))

	response := send("/u/{usuario.ID}/folders", data)
	bData, err := ioutil.ReadAll(response.Body)
	chk(err)

	var rData types.Folder
	err = json.Unmarshal(bData, &rData)
	chk(err)

	if rData.Folders != nil {
		fmt.Println("\nLa carpeta con nombre " + rData.Name + " se ha creado correctamente.")
		return true
	}
	fmt.Printf("Error al crear la carpeta: %v\n\n", rData)
	return false
}

func borrarCarpeta(deleteFolder string) bool {
	if deleteFolder != usuario.MainFolder {
		fmt.Printf("\nBorrando la carpeta con el nombre %s...", folder.Name)
		data := url.Values{}
		data.Set("user", crypto.Encode64([]byte(usuario.ID)))
		data.Set("token", crypto.Encode64([]byte(usuario.Token)))
		data.Set("folder", crypto.Encode64([]byte(deleteFolder)))

		response := send("/u/{usuario.ID}/folders/delete/{deleteFolder}", data)
		bData, err := ioutil.ReadAll(response.Body)
		chk(err)

		var rData types.Folder
		err = json.Unmarshal(bData, &rData)
		chk(err)

		if rData.Folders != nil {
			fmt.Println("\nLa carpeta con nombre " + rData.Name + " se ha borrado correctamente.")
			return true
		}
		fmt.Printf("\nError al borrar la carpeta: %v\n\n", rData)
		return false
	} else {
		fmt.Printf("\nNo se puede borrar la carpeta principal.\n")
		return false
	}
}

func exploredUnit(mainfolder string) {
	opt := ""
	i := 1
	match := false
	error := false
	var foldersIds map[int][]string
	foldersIds = make(map[int][]string)
	folderID := mainfolder
	folderName := "my-unit"
	for opt != "q" {
		i = 1
		foldersIds = make(map[int][]string)

		if !error {
			_ = getFolder(folderID)
			folderName = folder.Name
		}

		match = false
		if len(folder.Folders) != 0 {
			for key, value := range folder.Folders {
				fmt.Println(i, "- "+value+" ("+key+")")
				foldersIds[i] = []string{key, value}
				i = i + 1
			}
		} else {
			fmt.Println("-- No hay ningún archivo ni directorio. --")
		}

		fmt.Println("------------------------------------------")
		fmt.Printf("s - Subir fichero\n")
		fmt.Printf("c - Crear carpeta\n")
		if folderName != "my-unit" {
			fmt.Printf("b - Borrar carpeta\n")
			fmt.Printf("v - Volver\n")
		}
		fmt.Printf("q - Salir\n")
		fmt.Printf("Opcion: ")
		fmt.Scanf("%s\n", &opt)

		if opt != "q" && opt != "s" && opt != "v" && opt != "c" && opt != "b" {
			iter, err := strconv.Atoi(opt)
			if err != nil {
				fmt.Println("\nDebes introducir un número de la lista o q, ha introducido " + opt)
			} else {
				for key, value := range foldersIds {
					if key == iter {
						i = 1
						match = true
						folderID = value[0]
						folderName = value[1]
						error = false
					}
				}
				if !match {
					fmt.Println("\nLa opción introducida no existe, debe escoger de entre la lista\n")
					i = 1
					error = true
				}
			}
		} else {
			switch opt {
			case "s":
				uploadFile()
			case "v":
				folderID = folder.FolderParent
			case "q":
				fmt.Printf("\nBienvenido a su espacio personal " + usuario.Name + "\n\n")
			case "c":
				if crearCarpeta(folderID) {
					error = false
				}
			case "b":
				if borrarCarpeta(folderID) {
					folderID = folder.FolderParent
					error = false
				}
			}
		}
	}
}

func loggedMenu() {
	yellow := color.New(color.FgHiYellow).PrintfFunc()
	yellow("\nBienvenido a su espacio personal " + usuario.Name + ".\n")
	yellow("---------------------------------------------------\n")

	opt := ""
	for opt != "q" {
		color.Set(color.FgBlue)
		fmt.Printf("1 - Explorar mi espacio\nl - Logout\nq - Salir\nOpcion: ")
		fmt.Scanf("%s\n", &opt)
		color.Unset()
		switch opt {
		case "1":
			exploredUnit(usuario.MainFolder)
		case "l":
			fmt.Println("\nCerrando sesión...\n")
			usuario = types.User{}
			opt = "q"
		case "q":
			fmt.Println("\nHasta la próxima " + usuario.Name + "\n")
		default:
			fmt.Println("\nIntoduzca una opción correcta")
		}
	}
}

/**
Solicita el fichero que quiere subir el usuario
Trocea el fichero en bloques de 1MB
Y genera el fichero para el servidor
*/
func uploadFile() bool {
	fmt.Printf("Indique el fichero: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	path := scanner.Text()
	path = strings.Replace(path, "\\", "", -1)
	tokens := strings.Split(path, "/")
	fileName := tokens[len(tokens)-1]
	fmt.Println()
	file, err := os.Open(path)
	chk(err)
	defer file.Close()
	fileInfo, _ := file.Stat()
	var fileSize int64 = fileInfo.Size()
	const fileChunk = 1 * (1 << 20) // 1 MB
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	version := types.Version{
		ID: types.GenXid(),
	}
	fmt.Println("Subiendo archivo...")
	for i := uint64(0); i < totalPartsNum; i++ {
		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)
		file.Read(partBuffer)

		blockID := checkBlock(partBuffer) // Obtiene el id del bloque en el servidor
		version.Blocks = append(version.Blocks, blockID)
	}
	fileT := types.File{
		FolderID: folder.ID,
		Name:     fileName,
		OwnerID:  usuario.ID,
		Versions: append(make([]types.Version, 0), version),
	}
	return uploadFileT(fileT)
}

/**
Sube el tipo fichero al servidor
*/
func uploadFileT(file types.File) bool {
	data := url.Values{}
	fileB, err := json.Marshal(file)
	chk(err)
	data.Set("file", crypto.Encode64(fileB))
	data.Set("user", crypto.Encode64([]byte(usuario.ID)))
	data.Set("folder", crypto.Encode64([]byte(folder.ID)))
	data.Set("token", crypto.Encode64([]byte(usuario.Token)))

	resp := send("/u/"+usuario.ID+"/folders/"+folder.ID+"/upload", data)
	bData, err := ioutil.ReadAll(resp.Body)
	chk(err)
	response := types.Response{}
	err = json.Unmarshal(bData, &response)
	chk(err)
	if !response.Status {
		fmt.Println("Error: " + response.Msg)
		return false
	}
	fmt.Println("Archivo subido con éxito")
	return true
}

/**
Comprueba si existe el bloque en el seridor enviando el hash
Si no existe lo manda
Devuelve el identificador del bloque
*/

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
		_ = writer.WriteField("userID", crypto.Encode64([]byte(usuario.ID)))
		_ = writer.WriteField("folderID", crypto.Encode64([]byte(folder.ID)))
		_ = writer.WriteField("blockID", crypto.Encode64([]byte(blockID)))
		err = writer.Close()
		chk(err)
		req, err := http.NewRequest("POST", baseURL+"/uploadBlock", body)
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

// RunClient : run sincronice client
func main() {
	loadData()
	defer saveData()
	createClient()

	color.Yellow("\n===================================================")
	color.Yellow("============= Bienvenido a SincroNice =============")
	color.Yellow("===================================================\n")

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
			color.Set(color.FgBlue)
			fmt.Printf("1 - Login\n2 - Registro\nq - Salir\nOpcion: ")
			fmt.Scanf("%s\n", &opt)
			color.Unset()
		}

		switch opt {
		case "1":
			logged = login()
		case "2":
			registry()
		case "q":
			color.Green("Adios, gracias por usarnos.")
		default:
			color.Red("Intoduzca una opción correcta")
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
