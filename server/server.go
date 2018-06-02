package main

import (
	"SincroNice/crypto"
	"SincroNice/types"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

var (
	users   map[string]types.User
	folders map[string]types.Folder
	files   map[string]types.File
	blocks  map[string]types.Block
	port    = "8081"
)

const uploadPath = "C:/Users/pedro/go/src/SincroNice/server/tmp/"

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
	newPath := uploadPath + blockID
	newBlock, err := os.Create(newPath)
	defer newBlock.Close()
	chk(err)
	_, err = newBlock.Write(blockBytes)
	newBlock.Sync()
	chk(err)
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
			ID:     newFile.Versions[0].ID,
			Blocks: newFile.Versions[0].Blocks}
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
	router.HandleFunc("/uploadDrive", uploadDriveHandler)
	router.HandleFunc("/checkBlock", checkBlockHandler)
	router.HandleFunc("/u/{id}/folders/{folderId}/upload", uploadFile)
	router.HandleFunc("/u/{id}/folders", createFolder)
	router.HandleFunc("/u/{id}/folders/delete/{folderId}", deleteFolder)

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
