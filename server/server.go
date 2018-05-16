package main

import (

	// "context"
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

	"github.com/gorilla/mux"
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
const uploadPath = "./tmp"

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

	user, exist := users[userID]
	if !exist {
		r.Status = false
		r.Msg = "El usuario al que se intenta acceder no existe."
		response(w, r)
		log.Printf("Fail access to user %s", user.Email)
	} else {
		folder, exist := folders[folderID]
		if !exist {
			r.Status = false
			r.Msg = "El usuario " + user.Email + " no tiene carpeta principal."
			response(w, r)
			log.Printf("Fail access to main folder of user %s", user.Email)
		} else {
			r.Status = true
			r.Msg = "La carpeta no se ha podido encontrar"
			json.NewEncoder(w).Encode(folder)
			log.Printf("The user %s has correctly accessed the folder %s", user.Email, folder.Name)
		}
	}
}

// RunServer : run sincronice server
func main() {
	loadData()
	defer saveData()

	//fs := http.FileServer(http.Dir(uploadPath))

	log.Println("Running server on port: " + port)
	// suscripción SIGINT
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/register", registerHandler)
	router.HandleFunc("/checkToken", checkTokenHandler)
	router.HandleFunc("/u/{id}/my-unit", getMainFolder)
	router.HandleFunc("/u/{id}/folders/{folderId}", getFolder)

	srv := &http.Server{Addr: ":" + port, Handler: router}

	// metodo concurrente
	go func() {
		if err := srv.ListenAndServeTLS("server.crt", "server.key"); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan // espera señal SIGINT
	log.Println("\n\nShutdown server...")

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
	log.Println("Data saved")
}

//https://astaxie.gitbooks.io/build-web-application-with-golang/en/04.5.html

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
