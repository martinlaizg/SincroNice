package main

import (
	"SincroNice/types"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
)

var port = "8081"

var users map[string]types.User

func chk(e error) {
	if e != nil {
		panic(e)
	}
}

// response : recibe un objeto de un struct para responder al cliente
func response(w io.Writer, m interface{}) {
	rJSON, err := json.Marshal(&m) // codificamos en JSON
	chk(err)                       // comprobamos error
	w.Write(rJSON)                 // escribimos el JSON resultante
}

// RunServer : run sincronice server
func main() {
	loadData()
	defer saveData()

	log.Println("Running server on port: " + port)
	// suscripción SIGINT
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/register", registerHandler)
	router.HandleFunc("/u/{userID}/my-unit", registerHandler)

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
	/*
		ctx, fnc := context.WithTimeout(context.Background(), 5*time.Second)
		fnc()
		srv.Shutdown(ctx)
	*/
	log.Println("Servidor detenido correctamente")
}

func loadData() {
	log.Println("Loading data from JSON...")
	raw, err := ioutil.ReadFile("./db/users.json")
	chk(err)
	err = json.Unmarshal(raw, &users)
	chk(err)
	log.Println("Data loaded")
}

func saveData() {
	log.Println("Saving data to JSON...")
	raw, err := json.Marshal(users)
	chk(err)
	err = ioutil.WriteFile("./db/users.json", raw, 0777)
	chk(err)
	log.Println("Data saved")
}
