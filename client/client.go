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
		color.Red("error writing to buffer")
	}

	// open file handle
	fh, err := os.Open(ruta)
	if err != nil {
		color.Red("error opening file")
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
		color.Green("Subido correctamente\n")
		return
	}
	color.Red("Error al subir el archivo: %v\n", rData.Msg)
}

func login() bool {
	color.Set(color.FgHiYellow)
	fmt.Print("\nLogin\n")
	fmt.Print("Email: ")
	var email string
	fmt.Scanln(&email)
	fmt.Print("Password: ")
	bpass, err := gopass.GetPasswdMasked()
	chk(err)
	color.Unset()

	color.Yellow("\nAcceso como %s...\n", email)

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
		color.Red(rData.Msg)
		return false
	}
	usuario = rData.User
	return solicitarToken()
}

func solicitarToken() bool {
	color.Set(color.FgHiYellow)
	fmt.Println("Introduzca el token que le hemos enviado por correo electrónico")
	fmt.Print("Token: ")
	var token string
	fmt.Scanln(&token)
	data := url.Values{}
	color.Unset()

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
		color.Green("Sesión verificada correctamente")
	} else {
		usuario = types.User{}
		color.Red("El token introducido no coincide")
		return false
	}
	return true
}

func registry() {
	color.Set(color.FgHiYellow)
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
	color.Unset()

	color.Yellow("\nRegistrandose como %v... \n", email)
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
		color.Green("Registrado correctamente\n\n")
		return
	}
	color.Red("Error al registrarse: %v\n\n", resp.Msg)
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
		color.Magenta("\nSe encuentra en el directorio " + rData.Name + "\n")
		folder = rData
		return true
	}
	color.Red("Error al recuperar la carpeta: %v\n\n", rData)
	return false
}

func crearCarpeta(actualFolder string) bool {
	color.Set(color.FgHiYellow)
	fmt.Printf("Introduzca el nombre de la carpeta: ")
	color.Unset()
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	folderName := scanner.Text()

	color.Yellow("\nCreando la carpeta con el nombre %s...", folderName)
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
		color.Green("\nLa carpeta con nombre " + rData.Name + " se ha creado correctamente.")
		return true
	}
	color.Red("Error al crear la carpeta: %v\n\n", rData)
	return false
}

func borrarCarpeta(deleteFolder string) bool {
	if deleteFolder != usuario.MainFolder {
		color.Yellow("\nBorrando la carpeta con el nombre %s...", folder.Name)
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
			color.Green("\nLa carpeta con nombre " + rData.Name + " se ha borrado correctamente.")
			return true
		}
		color.Red("\nError al borrar la carpeta: %v\n\n", rData)
		return false
	} else {
		color.Red("\nNo se puede borrar la carpeta principal.\n")
		return false
	}
}

func exploredUnit(mainfolder string) {
	opt := ""
	i := 1
	match := false
	error := false
	var foldersIds map[int][]string
	var filesIds map[int][]string
	foldersIds = make(map[int][]string)
	filesIds = make(map[int][]string)
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
			color.Set(color.FgBlue)
			for key, value := range folder.Folders {
				fmt.Println(i, "- "+value+" ("+key+")")
				foldersIds[i] = []string{key, value}
				i = i + 1
			}
			color.Unset()
			if len(folder.Files) != 0 {
				for key, value := range folder.Files {
					fmt.Println(i, "- "+value+" ("+key+")")
					filesIds[i] = []string{key, value}
					i = i + 1
				}
			}
		} else {
			color.Red("-- No hay ningún archivo ni directorio. --")
		}

		color.Set(color.FgHiYellow)
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
		color.Unset()

		if opt != "q" && opt != "s" && opt != "v" && opt != "c" && opt != "b" {
			iter, err := strconv.Atoi(opt)
			if err != nil {
				color.Red("\nDebes introducir un número de la lista o q, ha introducido " + opt)
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
					color.Red("\nLa opción introducida no existe, debe escoger de entre la lista\n")
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
				color.Magenta("\nBienvenido a su espacio personal, " + usuario.Name + ".\n")
				color.Magenta("---------------------------------------------------\n")
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
	color.Magenta("\nBienvenido a su espacio personal, " + usuario.Name + ".\n")
	color.Magenta("---------------------------------------------------\n")

	opt := ""
	for opt != "q" {
		color.Set(color.FgHiYellow)
		fmt.Printf("1 - Explorar mi espacio\nl - Logout\nq - Salir\nOpcion: ")
		fmt.Scanf("%s\n", &opt)
		color.Unset()
		switch opt {
		case "1":
			exploredUnit(usuario.MainFolder)
		case "l":
			color.Yellow("\nCerrando sesión...\n")
			usuario = types.User{}
			opt = "q"
		case "q":
			color.Yellow("\nHasta la próxima " + usuario.Name + "\n")
		default:
			color.Red("\nIntoduzca una opción correcta")
		}
	}
}

/**
Solicita el fichero que quiere subir el usuario
Trocea el fichero en bloques de 1MB
Y genera el fichero para el servidor
*/
func uploadFile() bool {
	color.Yellow("Indique el fichero: ")

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
	color.Yellow("Subiendo archivo...")
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
		color.Red("Error: " + response.Msg)
		return false
	}
	color.Green("Archivo subido con éxito")
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

	logged := false
	error := false

	color.Yellow("\n===================================================")
	color.Yellow("============= Bienvenido a SincroNice =============")
	color.Yellow("===================================================\n")

	for opt := ""; opt != "q"; {
		if usuario.Token != "" && !error {
			logged = true
		}

		if logged {
			loggedMenu()
			logged = false
		}
		if !logged {
			error = false
			color.Set(color.FgHiYellow)
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
			color.Green("Adios, gracias por usar nuestros servicios.")
		default:
			color.Red("\nIntoduzca una opción correcta")
			error = true
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
