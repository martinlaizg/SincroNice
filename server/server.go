package main

import (
	"SincroNice/types"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var port = "8081"

func chk(e error) {
	if e != nil {
		panic(e)
	}
}

func response(w io.Writer, status bool, msg string) {
	r := types.Resp{Status: status, Msg: msg} // formateamos respuesta
	rJSON, err := json.Marshal(&r)            // codificamos en JSON
	chk(err)                                  // comprobamos error
	w.Write(rJSON)                            // escribimos el JSON resultante
}

func getMux() (mux *http.ServeMux) {
	mux = http.NewServeMux()

	//mux.Handle("/", http.HandlerFunc(handler))
	mux.Handle("/login", http.HandlerFunc(loginHandler))

	return
}

// RunServer : run sincronice server
func main() {
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
	log.Println("Shutdown server...")

	// apagar servidor de forma segura
	// ctx, fnc := context.WithTimeout(context.Background(), 5*time.Second)
	// fnc()
	// srv.Shutdown(ctx)
	// log.Println("Servidor detenido correctamente")
}
