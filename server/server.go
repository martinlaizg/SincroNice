package main

import (
	"SincroNice/crypto"
	"SincroNice/types"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v2"
)

var (
	users   map[string]types.User
	folders map[string]types.Folder
	files   map[string]types.File
	blocks  map[string]types.Block
	port    = "8081"
)

func chk(e error) {
	if e != nil {
		panic(e)
	}
}

const maxUploadSize = 2 * 1024 // 2 MB
const uploadPath = "./tmp/"

// response : recibe un objeto de un struct para responder al cliente
func response(w io.Writer, m interface{}) {
	rJSON, err := json.Marshal(&m) // codificamos en JSON
	chk(err)                       // comprobamos error
	w.Write(rJSON)                 // escribimos el JSON resultante

}

func getMainFolder(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")

	userID := string(crypto.Decode64(req.Form.Get("id")))
	user, exist := users[userID]
	if !exist {
		r.Status = false
		r.Msg = "El usuario al que se intenta acceder no existe."
		response(w, r)
		log.Printf("Fail access to user %s", user.Email)
	} else {
		folder, exist := folders[user.MainFolder]
		if !exist {
			r.Status = false
			r.Msg = "El usuario " + user.Email + " no tiene carpeta principal."
			response(w, r)
			log.Printf("Fail access to main folder of user %s", user.Email)
		} else {
			r.Status = true
			r.Msg = "La carpeta principal del usuario se ha encontrado"
			json.NewEncoder(w).Encode(folder)
			log.Printf("The user %s has correctly accessed the folder %s", user.Email, folder.Name)
		}
	}
}

func getFolder(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")

	userID := string(crypto.Decode64(req.Form.Get("id")))
	folderID := string(crypto.Decode64(req.Form.Get("folderId")))
	token := string(crypto.Decode64(req.Form.Get("token")))

	user, exist := users[userID]
	if !exist {
		r.Status = false
		r.Msg = "El usuario al que se intenta acceder no existe."
		response(w, r)
		log.Printf("Fail access to user %s", user.Email)
	} else {
		if chkToken(token, userID) {
			folder, exist := folders[folderID]
			if !exist {
				r.Status = false
				r.Msg = "El usuario " + user.Email + " no tiene carpeta principal."
				response(w, r)
				log.Printf("Fail access to main folder of user %s", user.Email)
			} else {
				r.Status = true
				r.Msg = "Hemos encontrado la carpeta"
				json.NewEncoder(w).Encode(folder)
				log.Printf("The user %s has correctly accessed the folder %s", user.Email, folder.Name)
			}
		} else {
			r.Status = false
			r.Msg = "El token utilizado no es correcto."
			response(w, r)
			log.Printf("Fail access to user %s", user.Email)
		}
	}
}

func deleteSubFolders(subFolders map[string]string) {
	for key, value := range folders {
		for key2 := range subFolders {
			if key == key2 {
				if len(value.Folders) != 0 {
					deleteSubFolders(value.Folders)
				} else {
					delete(folders, key)
				}
				delete(folders, key)
			}
		}
	}
}

func deleteFile(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")

	userID := string(crypto.Decode64(req.Form.Get("user")))
	fileID := string(crypto.Decode64(req.Form.Get("file")))
	token := string(crypto.Decode64(req.Form.Get("token")))

	user, exist := users[userID]
	if !exist {
		r.Status = false
		r.Msg = "El usuario al que se intenta acceder no existe."
		response(w, r)
		log.Printf("Fail access to user %s", user.Email)
	} else {
		if chkToken(token, userID) {
			file, exist := files[fileID]
			if !exist {
				r.Status = false
				r.Msg = "El usuario " + user.Email + " no tiene carpeta principal."
				response(w, r)
				log.Printf("Fail access to main folder of user %s", user.Email)
			} else {
				if file.OwnerID == user.ID {
					r.Status = true
					r.Msg = "Hemos eliminado el archivo"
					delete(files, file.ID)
					for _, folder := range folders {
						for id := range folder.Files {
							if id == fileID {
								delete(folder.Files, fileID)
							}
						}
					}
					json.NewEncoder(w).Encode(file)
					log.Printf("The user %s has correctly deleted the file %s", user.Email, file.Name)
				} else {
					r.Status = false
					r.Msg = "No tienes permiso para eliminar el archivo"
					response(w, r)
					log.Printf("Fail access to user %s to file %s", user.Email, file.Name)
				}
			}
		}
	}
}

func deleteFolder(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")

	userID := string(crypto.Decode64(req.Form.Get("user")))
	folderID := string(crypto.Decode64(req.Form.Get("folder")))
	token := string(crypto.Decode64(req.Form.Get("token")))

	user, exist := users[userID]
	if !exist {
		r.Status = false
		r.Msg = "El usuario al que se intenta acceder no existe."
		response(w, r)
		log.Printf("Fail access to user %s", user.Email)
	} else {
		if chkToken(token, userID) {
			folder, exist := folders[folderID]
			if !exist {
				r.Status = false
				r.Msg = "El usuario " + user.Email + " no tiene carpeta principal."
				response(w, r)
				log.Printf("Fail access to main folder of user %s", user.Email)
			} else {
				r.Status = true
				r.Msg = "Hemos eliminado la carpeta"
				for key, value := range folders {
					if key == folderID {
						deleteSubFolders(folder.Folders)
						delete(folders, key)
					}
					delete(value.Folders, folderID)
				}
				json.NewEncoder(w).Encode(folder)
				log.Printf("The user %s has correctly deleted the folder %s", user.Email, folder.Name)
			}
		} else {
			r.Status = false
			r.Msg = "El token utilizado no es correcto."
			response(w, r)
			log.Printf("Fail access to user %s", user.Email)
		}
	}
}

func createFolder(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")

	userID := string(crypto.Decode64(req.Form.Get("user")))
	folderName := string(crypto.Decode64(req.Form.Get("folderName")))
	actualFolder := string(crypto.Decode64(req.Form.Get("actualFolder")))
	token := string(crypto.Decode64(req.Form.Get("token")))

	user, exist := users[userID]
	if !exist {
		r.Status = false
		r.Msg = "El usuario al que se intenta acceder no existe."
		response(w, r)
		log.Printf("Fail access to user %s", user.Email)
	} else {
		if chkToken(token, userID) {
			folder, exist := folders[actualFolder]
			if !exist {
				r.Status = false
				r.Msg = "El usuario " + user.Email + " no tiene carpeta principal."
				response(w, r)
				log.Printf("Fail access to main folder of user %s", user.Email)
			} else {
				folderID := types.GenXid()
				folder.Folders[folderID] = folderName
				folder := types.Folder{
					ID:           folderID,
					UserID:       userID,
					Name:         folderName,
					Path:         "/",
					Created:      time.Now().UTC().String(),
					Updated:      time.Now().UTC().String(),
					FolderParent: folder.ID,
					Folders:      make(map[string]string),
					Files:        make(map[string]string)}
				folders[folderID] = folder
				r.Status = true
				r.Msg = "La carpeta ha sido creada correctamente"
				json.NewEncoder(w).Encode(folder)
				log.Printf("The user %s has correctly created the folder %s", user.Email, folder.Name)
			}
		} else {
			r.Status = false
			r.Msg = "El token utilizado no es correcto."
			response(w, r)
			log.Printf("Fail access to user %s", user.Email)
		}
	}
}

func checkBlock(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")
	hash := crypto.Decode64(req.Form.Get("hash"))
	r.Status = false
	r.Msg = types.GenXid()
	for _, block := range blocks {
		if string(block.Hash) == string(hash) {
			r.Status = true
			r.Msg = block.ID
		}
	}
	response(w, r)
}

/**
Inserta el bloque en la base de datos y lo almacena en el sistema
*/
func uploadBlock(w http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(1)
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")
	blockID := string(crypto.Decode64(req.PostFormValue("blockID")))
	userID := string(crypto.Decode64(req.PostFormValue("userID")))
	block, _, err := req.FormFile("fileupload") // Obtenemos el fichero
	defer block.Close()
	chk(err)
	blockBytes, err := ioutil.ReadAll(block) // Lo pasamos a bytes
	chk(err)
	hash := crypto.Hash(blockBytes)
	blockT := types.Block{
		ID:    blockID,
		Hash:  hash[:],
		Owner: userID}
	blocks[blockT.ID] = blockT

	//newPath := uploadPath + blockID
	//newBlock, err := os.Create(newPath)
	//defer newBlock.Close()
	//chk(err)
	//_, err = newBlock.Write(blockBytes)
	//newBlock.Sync()
	//chk(err)

	//Subimos al drive
	ctx := context.Background()
	// process the credential file
	credential, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(credential, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(ctx, config)

	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}

	token, err := tokenFromFile(cacheFile)
	if err != nil {
		log.Fatalf("Unable to get token from file. %v", err)
	}

	fileMIMEType := http.DetectContentType(blockBytes)

	postURL := "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart"
	authToken := token.AccessToken

	boundary := randStr(32, "alphanum")

	uploadData := []byte("\n" +
		"--" + boundary + "\n" +
		"Content-Type: application/json; charset=" + string('"') + "UTF-8" + string('"') + "\n\n" +
		"{ \n" +
		string('"') + "name" + string('"') + ":" + string('"') + blockID + string('"') + "\n" +
		"} \n\n" +
		"--" + boundary + "\n" +
		"Content-Type:" + fileMIMEType + "\n\n" +
		string(blockBytes) + "\n" +

		"--" + boundary + "--")

	// post to Drive with RESTful method
	request, _ := http.NewRequest("POST", postURL, strings.NewReader(string(uploadData)))
	request.Header.Add("Host", "www.googleapis.com")
	request.Header.Add("Authorization", "Bearer "+authToken)
	request.Header.Add("Content-Type", "multipart/related; boundary="+string('"')+boundary+string('"'))
	request.Header.Add("Content-Length", strconv.FormatInt(request.ContentLength, 10))

	response2, err := client.Do(request)

	body, err := ioutil.ReadAll(response2.Body)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
	}

	fmt.Printf(string(body))

	if err != nil {
		log.Fatalf("Unable to be post to Google API: %v", err)
		r.Status = false
		response(w, r)
		return
	}
	defer response2.Body.Close()

	r.Status = true
	response(w, r)
}

/**
Inserta el tipo de fichero en la base de datos
*/
func uploadFile(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")
	newFile := types.File{}
	userID := string(crypto.Decode64(req.Form.Get("user")))
	folderID := string(crypto.Decode64(req.Form.Get("folder")))
	token := string(crypto.Decode64(req.Form.Get("token")))
	log.Println("Usuario " + userID + " intentó subir fichero")
	if !chkToken(token, userID) { // Verificamos la validez del usuario
		r.Status = false
		r.Msg = "No está bien autenticado"
		response(w, r)
		log.Println("Token de usuario no verificado")
		return
	}
	err := json.Unmarshal(crypto.Decode64(req.Form.Get("file")), &newFile)
	chk(err)
	folder, exist := folders[folderID]
	if !exist || folder.UserID != userID || newFile.OwnerID != userID { // Carpeta incorrecta
		r.Status = false
		r.Msg = "Carpeta incorrecta"
		response(w, r)
		return
	}
	r.Status = true
	r.Msg = "Subido correctamente"

	fileID := ""
	for id, file := range folder.Files {
		if file == newFile.Name {
			fileID = id
		}
	}
	file, exist := files[fileID]
	if !exist {
		newFile.ID = types.GenXid()
		files[newFile.ID] = newFile
		folders[folderID].Files[newFile.ID] = newFile.Name
		log.Println("Creado nuevo fichero")
	} else {
		newVersion := types.Version{
			ID:      newFile.Versions[0].ID,
			Created: time.Now().UTC().String(),
			Blocks:  newFile.Versions[0].Blocks}
		file.Versions = append(file.Versions, newVersion)
		files[fileID] = file
		log.Println("Añadida nueva versión al fichero ya existente")
	}

	response(w, r)
}

// RunServer : run sincronice server
func main() {
	log.Printf("Servidor a la espera de peticiones.")
	f, err := os.OpenFile("LogFile", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Printf("Running server...")
	loadData()
	defer saveData()

	log.Printf("Running server on port: " + port)
	// suscripción SIGINT
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/register", registerHandler)
	router.HandleFunc("/checkToken", checkTokenHandler)
	router.HandleFunc("/checkBlock", checkBlock)
	router.HandleFunc("/uploadBlock", uploadBlock)
	router.HandleFunc("/u/{id}/my-unit", getMainFolder)
	router.HandleFunc("/u/{id}/folders/{folderId}", getFolder)
	router.HandleFunc("/u/{id}/folders/{folderId}/upload", uploadFile)
	router.HandleFunc("/u/{id}/folders", createFolder)
	router.HandleFunc("/u/{id}/folders/delete/{folderId}", deleteFolder)
	router.HandleFunc("/u/{id}/files/{fileID}", deleteFile)

	srv := &http.Server{Addr: ":" + port, Handler: router}

	// metodo concurrente
	go func() {
		if err := srv.ListenAndServeTLS("server.crt", "server.key"); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan // espera señal SIGINT
	log.Printf("\n\nShutdown server...")

	// apagar servidor de forma segura
	// ctx, fnc := context.WithTimeout(context.Background(), 5*time.Second)
	// fnc()
	// srv.Shutdown(ctx)
	log.Println("Servidor detenido correctamente")
}

func loadData() {
	log.Println("Loading data from JSON...")
	raw, err := ioutil.ReadFile("./db/users.json")
	chk(err)
	err = json.Unmarshal(raw, &users)
	chk(err)
	raw, err = ioutil.ReadFile("./db/folders.json")
	chk(err)
	err = json.Unmarshal(raw, &folders)
	chk(err)
	raw, err = ioutil.ReadFile("./db/files.json")
	chk(err)
	err = json.Unmarshal(raw, &files)
	chk(err)
	raw, err = ioutil.ReadFile("./db/blocks.json")
	chk(err)
	err = json.Unmarshal(raw, &blocks)
	chk(err)
	log.Println("Data loaded")
}

func saveData() {
	log.Println("Saving data to JSON...")
	raw, err := json.Marshal(users)
	chk(err)
	err = ioutil.WriteFile("./db/users.json", raw, 0777)
	chk(err)
	raw, err = json.Marshal(folders)
	chk(err)
	err = ioutil.WriteFile("./db/folders.json", raw, 0777)
	chk(err)
	raw, err = json.Marshal(files)
	chk(err)
	err = ioutil.WriteFile("./db/files.json", raw, 0777)
	chk(err)
	raw, err = json.Marshal(blocks)
	chk(err)
	err = ioutil.WriteFile("./db/blocks.json", raw, 0777)
	chk(err)
	log.Println("Data saved")
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func uploadDriveHandler(w http.ResponseWriter, req *http.Request) {

	req.ParseMultipartForm(1)
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")
	blockID := string(crypto.Decode64(req.PostFormValue("blockID")))
	userID := string(crypto.Decode64(req.PostFormValue("userID")))
	block, _, err := req.FormFile("fileupload") // Obtenemos el fichero
	defer block.Close()
	chk(err)

	blockBytes, err := ioutil.ReadAll(block) // Lo pasamos a bytes
	chk(err)

	hash := crypto.Hash(blockBytes)

	blockT := types.Block{
		ID:    blockID,
		Hash:  hash[:],
		Owner: userID,
	}
	blocks[blockT.ID] = blockT

	r.Status = true
	response(w, r)

	//upload drive
	ctx := context.Background()
	// process the credential file
	credential, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(credential, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(ctx, config)

	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}

	token, err := tokenFromFile(cacheFile)
	if err != nil {
		log.Fatalf("Unable to get token from file. %v", err)
	}

	//id := "1SXfqr0Jm6iEe04W5BGvo2X57pYvatDjY"
	//DownloadFile(client, id)

	/*fileBytes, err := ioutil.ReadFile(ruta + blockID)
	if err != nil {
		log.Fatalf("Unable to read file for upload: %v", err)
	}*/

	fileMIMEType := http.DetectContentType(blockBytes)

	postURL := "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart"
	authToken := token.AccessToken

	boundary := randStr(32, "alphanum")

	uploadData := []byte("\n" +
		"--" + boundary + "\n" +
		"Content-Type: application/json; charset=" + string('"') + "UTF-8" + string('"') + "\n\n" +
		"{ \n" +
		string('"') + "name" + string('"') + ":" + string('"') + blockID + string('"') + "\n" +
		"} \n\n" +
		"--" + boundary + "\n" +
		"Content-Type:" + fileMIMEType + "\n\n" +
		string(blockBytes) + "\n" +

		"--" + boundary + "--")

	// post to Drive with RESTful method
	request, _ := http.NewRequest("POST", postURL, strings.NewReader(string(uploadData)))
	request.Header.Add("Host", "www.googleapis.com")
	request.Header.Add("Authorization", "Bearer "+authToken)
	request.Header.Add("Content-Type", "multipart/related; boundary="+string('"')+boundary+string('"'))
	request.Header.Add("Content-Length", strconv.FormatInt(request.ContentLength, 10))

	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("Unable to be post to Google API: %v", err)
		return
	}

	defer response.Body.Close()
	//body, err := ioutil.ReadAll(response.Body)

	/*if err != nil {
		log.Fatalf("Unable to read Google API response: %v", err)
		return
	}*/

	//	fmt.Println(string(body))

	log.Println("File " + blockID + " upload successful")
}

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("google-drive-golang.json")), err
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func randStr(strSize int, randType string) string {

	var dictionary string

	if randType == "alphanum" {
		dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "alpha" {
		dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "number" {
		dictionary = "0123456789"
	}

	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}
