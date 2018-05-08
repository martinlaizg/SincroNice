package main

import (
	"SincroNice/types"
	"crypto/rand"
	"fmt"
	//"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	//"time"
)

var (
	users   map[string]types.User
	folders map[string]types.Folder
	files   map[string]types.File
	port    = "8081"
)

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

const maxUploadSize = 2 * 1024 // 2 MB
const uploadPath = "./tmp"

func getMux() (mux *http.ServeMux) {
	mux = http.NewServeMux()

	mux.Handle("/login", http.HandlerFunc(loginHandler))
	mux.Handle("/register", http.HandlerFunc(registerHandler))
	mux.Handle("/upload", http.HandlerFunc(uploadHandler))

	return
}

// RunServer : run sincronice server
func main() {
	loadData()
	defer saveData()

	fs := http.FileServer(http.Dir(uploadPath))
	http.Handle("/files/", http.StripPrefix("/files", fs))

	log.Println("Running server on port: " + port)
	// suscripción SIGINT
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	mux := getMux()

	srv := &http.Server{Addr: ":" + port, Handler: mux}

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
