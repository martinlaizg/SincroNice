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
	"time"

	"github.com/fatih/color"
	"github.com/howeyc/gopass"
)

var baseURL = "https://localhost:8081"

var client *http.Client

var usuario types.User
var folder types.Folder
var file types.File

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
		color.Yellow("\n---------------------------------------------------")
		color.Magenta("Se encuentra en el directorio " + rData.Name)
		color.Yellow("---------------------------------------------------\n")
		folder = rData
		return true
	}
	color.Red("Error al recuperar la carpeta: %v\n\n", rData)
	return false
}

func getFile(id string) bool {
	data := url.Values{}
	data.Set("id", crypto.Encode64([]byte(usuario.ID)))
	data.Set("token", crypto.Encode64([]byte(usuario.Token)))
	data.Set("fileID", crypto.Encode64([]byte(id)))

	response := send("/u/{usuario.ID}/files/{id}", data)
	bData, err := ioutil.ReadAll(response.Body)
	chk(err)
	var rData types.File
	err = json.Unmarshal(bData, &rData)
	chk(err)

	if rData.Versions != nil {
		file = rData
		return true
	}
	color.Red("Error al recuperar el archivo: %v\n\n", rData)
	return false
}

func crearCarpeta(actualFolder string) bool {
	color.Set(color.FgYellow)
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
		color.Green("La carpeta con nombre " + rData.Name + " se ha creado correctamente.")
		return true
	}
	color.Yellow("\n---------------------------------------------------")
	color.Red("Error al crear la carpeta: %v\n\n", rData)
	color.Yellow("---------------------------------------------------\n")
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
			color.Yellow("\n---------------------------------------------------")
			color.Green("La carpeta con nombre " + rData.Name + " se ha borrado correctamente.")
			color.Yellow("---------------------------------------------------\n")
			return true
		}
		color.Yellow("\n---------------------------------------------------")
		color.Red("\nError al borrar la carpeta: %v", rData)
		color.Yellow("---------------------------------------------------\n")
		return false
	} else {
		color.Yellow("\n---------------------------------------------------")
		color.Red("No se puede borrar la carpeta principal.")
		color.Yellow("---------------------------------------------------\n")
		return false
	}
}

func deleteFile(id string, name string) bool {
	color.Yellow("\nBorrando el archivo con el nombre %s...", name)
	data := url.Values{}
	data.Set("user", crypto.Encode64([]byte(usuario.ID)))
	data.Set("token", crypto.Encode64([]byte(usuario.Token)))
	data.Set("file", crypto.Encode64([]byte(id)))

	response := send("/u/{usuario.ID}/files/delete/{id}", data)
	bData, err := ioutil.ReadAll(response.Body)
	chk(err)

	var rData types.File
	err = json.Unmarshal(bData, &rData)
	chk(err)

	if rData.Versions != nil {
		color.Yellow("\n---------------------------------------------------")
		color.Green("El archivo con el nombre " + rData.Name + " se ha borrado correctamente.")
		color.Yellow("---------------------------------------------------\n")
		return true
	}
	color.Yellow("\n---------------------------------------------------")
	color.Red("\nError al borrar el archivo: %v", rData)
	color.Yellow("---------------------------------------------------\n")
	return false
}

func downloadFile() bool {
	match := false
	opt := ""
	version := ""

	color.Set(color.FgYellow)
	fmt.Printf("\n---------------------------------------------------")
	fmt.Printf("\n¿Qué versión del archivo ")
	color.Unset()
	fmt.Printf(file.Name)
	color.Yellow(" quiere descargar?\n")
	color.Yellow("---------------------------------------------------")
	for opt != "q" && opt != "Q" {
		if len(file.Versions) != 0 {
			color.Set(color.FgHiBlue)
			for key, value := range file.Versions {
				fmt.Println(key+1, "- "+value.Created)
			}
			color.Unset()
		}
		color.Set(color.FgYellow)
		fmt.Println("---------------------------------------------------")
		fmt.Println("(Q)uit.")
		fmt.Printf("Opcion: ")
		color.Unset()
		fmt.Scanf("%s\n", &opt)
		if opt != "q" && opt != "Q" {
			iter, err := strconv.Atoi(opt)
			if err != nil {
				color.Yellow("\n---------------------------------------------------")
				color.Red("Debes introducir un número de la lista o Q, ha introducido " + opt)
				color.Yellow("---------------------------------------------------\n")
			} else {
				for key, value := range file.Versions {
					if key == iter-1 {
						match = true
						version = value.ID
					}
				}
				if !match {
					color.Yellow("\n---------------------------------------------------")
					color.Red("La opción introducida no existe, debe escoger de entre la lista")
					color.Yellow("---------------------------------------------------\n")
				} else {
					color.Yellow("\nDescargando el archivo con el nombre %s...", file.Name)
					data := url.Values{}
					data.Set("user", crypto.Encode64([]byte(usuario.ID)))
					data.Set("token", crypto.Encode64([]byte(usuario.Token)))
					data.Set("file", crypto.Encode64([]byte(file.ID)))
					data.Set("version", crypto.Encode64([]byte(version)))

					response := send("/u/{usuario.ID}/files/{name}/versions/{version}", data)
					bData, err := ioutil.ReadAll(response.Body)
					chk(err)

					var rData types.Version
					err = json.Unmarshal(bData, &rData)
					chk(err)

					if rData.ID == version {
						color.Yellow("\n---------------------------------------------------")
						color.Green("El archivo con el nombre " + file.Name + " se ha descargado correctamente.")
						color.Yellow("---------------------------------------------------\n")
						return true
					}
					color.Yellow("\n---------------------------------------------------")
					color.Red("\nError al borrar el archivoasdasdas: %v", rData)
					color.Yellow("---------------------------------------------------\n")
					return false
				}
			}
		}
		return false
	}
	return false
}

func exploredUnit(mainfolder string) {
	opt := ""
	i := 1
	match := false
	matchFile := false
	error := false
	var foldersIds map[int][]string
	var filesIds map[int][]string
	folderID := mainfolder
	folderName := "my-unit"
	fileID := ""
	fileName := ""
	for opt != "q" {
		i = 1

		foldersIds = make(map[int][]string)
		filesIds = make(map[int][]string)

		if !error {
			_ = getFolder(folderID)
			folderName = folder.Name
		}

		if len(folder.Folders) != 0 && !matchFile {
			color.Set(color.FgHiBlue)
			for key, value := range folder.Folders {
				fmt.Println(i, "- "+value+" ("+key+")")
				foldersIds[i] = []string{key, value}
				i = i + 1
			}
			color.Unset()
		}
		if len(folder.Files) != 0 && fileID == "" {
			for key, value := range folder.Files {
				fmt.Println(i, "- "+value+" ("+key+")")
				filesIds[i] = []string{key, value}
				i = i + 1
			}
		}
		if len(folder.Folders) == 0 && len(folder.Files) == 0 {
			color.Red("---- No hay directorios ni archivos ----")
		}
		if fileID == "" {
			color.Set(color.FgYellow)
			fmt.Println("---------------------------------------------------")
			fmt.Printf("s - Subir fichero\n")
			fmt.Printf("c - Crear carpeta\n")
			if folderName != "my-unit" {
				fmt.Printf("b - Borrar carpeta\n")
				fmt.Printf("v - Volver\n")
			}
			fmt.Printf("q - Salir\n")
			fmt.Printf("Opcion: ")
			color.Unset()
			fmt.Scanf("%s\n", &opt)

			if opt != "q" && opt != "s" && opt != "v" && opt != "c" && opt != "b" {
				iter, err := strconv.Atoi(opt)
				if err != nil {
					color.Yellow("\n---------------------------------------------------")
					color.Red("Debes introducir un número de la lista o q, ha introducido " + opt)
					color.Yellow("---------------------------------------------------\n")
				} else {
					match = false
					for key, value := range foldersIds {
						if key == iter {
							i = 1
							match = true
							matchFile = false
							folderID = value[0]
							folderName = value[1]
							error = false
						}
					}
					if !match {
						for key, value := range filesIds {
							if key == iter {
								i = 1
								matchFile = true
								fileID = value[0]
								fileName = value[1]
								error = true
							}
						}
						if !matchFile {
							color.Yellow("\n---------------------------------------------------")
							color.Red("La opción introducida no existe, debe escoger de entre la lista")
							color.Yellow("---------------------------------------------------\n")
							i = 1
							error = true
						}
					}
				}
			} else {
				switch opt {
				case "s":
					uploadFile()
				case "v":
					folderID = folder.FolderParent
				case "q":
					color.Yellow("\n---------------------------------------------------\n")
					color.Set(color.FgHiMagenta)
					fmt.Printf("Bienvenido a su espacio personal, " + usuario.Name + ".\n")
					color.Unset()
					color.Yellow("---------------------------------------------------\n")
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
		} else {
			if getFile(fileID) {
				fileMenu(fileID, fileName)
			}
			fileID = ""
			fileName = ""
			error = false
			matchFile = false
		}
	}
}

func fileMenu(id string, name string) {
	opt := ""

	for opt != "q" && opt != "Q" {
		color.Set(color.FgYellow)
		fmt.Printf("\n---------------------------------------------------")
		fmt.Printf("\n¿Qué desea hacer con el archvio ")
		color.Unset()
		fmt.Printf(name)
		color.Yellow("?\n")
		color.Set(color.FgYellow)
		fmt.Println("---------------------------------------------------")
		fmt.Println("(B)orrar archivo.")
		fmt.Println("(D)escargar archivo.")
		fmt.Println("(Q)uit.")
		fmt.Printf("Opcion: ")
		color.Unset()
		fmt.Scanf("%s\n", &opt)
		if opt == "b" || opt == "d" || opt == "B" || opt == "D" {
			opt = strings.ToUpper(opt)
			switch opt {
			case "D":
				downloadFile()
			case "B":
				deleteFile(id, name)
				opt = "q"
			}
		}
	}
}

func loggedMenu() {
	color.Yellow("\n---------------------------------------------------\n")
	color.Set(color.FgHiMagenta)
	fmt.Printf("Bienvenido a su espacio personal, " + usuario.Name + ".\n")
	color.Unset()
	color.Yellow("---------------------------------------------------\n")

	opt := ""
	for opt != "q" {
		color.Set(color.FgYellow)
		fmt.Printf("1 - Explorar mi espacio\nl - Logout\nq - Salir\nOpcion: ")
		color.Unset()
		fmt.Scanf("%s\n", &opt)

		switch opt {
		case "1":
			exploredUnit(usuario.MainFolder)
		case "l":
			color.Yellow("\nCerrando sesión...\n")
			usuario = types.User{}
			opt = "q"
		case "q":
			color.Yellow("\n---------------------------------------------------")
			color.Magenta("Hasta la próxima " + usuario.Name + "\n")
			color.Yellow("---------------------------------------------------\n")

		default:
			color.Yellow("\n---------------------------------------------------")
			color.Red("Intoduzca una opción correcta")
			color.Yellow("---------------------------------------------------\n")
		}
	}
}

/**
Solicita el fichero que quiere subir el usuario
Trocea el fichero en bloques de 1MB
Y genera el fichero para el servidor
*/
func uploadFile() bool {
	color.Set(color.FgYellow)
	fmt.Printf("Indique el fichero: ")
	color.Unset()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	path := scanner.Text()
	path = strings.Replace(path, "\\", "", -1)
	tokens := strings.Split(path, "/")
	fileName := tokens[len(tokens)-1]
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("No se pudo abrir el archivo " + path)
		return false
	}
	defer file.Close()
	fileInfo, _ := file.Stat()
	var fileSize int64 = fileInfo.Size()
	const fileChunk = 1 * (1 << 20) // 1 MB
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	version := types.Version{
		ID:      types.GenXid(),
		Created: time.Now().UTC().String(),
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
			color.Set(color.FgYellow)
			fmt.Printf("1 - Login\n2 - Registro\nq - Salir\nOpcion: ")
			color.Unset()
			fmt.Scanf("%s\n", &opt)
		}

		switch opt {
		case "1":
			logged = login()
		case "2":
			registry()
		case "q":
			color.Yellow("\n---------------------------------------------------")
			color.Magenta("Adios, gracias por usar nuestros servicios.")
			color.Yellow("---------------------------------------------------\n\n")
		default:
			color.Yellow("\n---------------------------------------------------")
			color.Red("Intoduzca una opción correcta")
			color.Yellow("---------------------------------------------------\n")
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
