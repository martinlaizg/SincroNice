package main

import (
	"SincroNice/types"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

var port = "8081"

var users map[string]types.User
var folders map[int]types.Folder
var files map[string]types.File

func chk(e error) {
	if e != nil {
		panic(e)
	}
}

func response(w io.Writer, status bool, msg string) {
	r := types.Response{Status: status, Msg: msg} // formateamos respuesta
	rJSON, err := json.Marshal(&r)                // codificamos en JSON
	chk(err)                                      // comprobamos error
	w.Write(rJSON)                                // escribimos el JSON resultante
}

func userShow(w http.ResponseWriter, req *http.Request) {
	log.Println("User try to get the folders and files.")
	req.ParseForm()
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(req)
	userID := vars["userId"]
	user, exists := users[userID]
	if !exists {
		response(w, false, "No existe el usuario")
		log.Println("Failed to search user files.")
		return
	}
	resp, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(resp)
	log.Println("It has been possible to find the user's space")
}

// RunServer : run sincronice server
func main() {
	loadData()

	log.Println("Running server on port: " + port)
	// suscripción SIGINT
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/register", registerHandler)
	router.HandleFunc("/u/{userId}", userShow)

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
	ctx, fnc := context.WithTimeout(context.Background(), 5*time.Second)
	fnc()
	srv.Shutdown(ctx)
	saveData()
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
