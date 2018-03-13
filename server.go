package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func chk(e error) {
	if e != nil {
		panic(e)
	}
}

type resp struct {
	Ok  bool   // true -> correcto, false -> error
	Msg string // mensaje adicional
}

func response(w io.Writer, ok bool, msg string) {
	r := resp{Ok: ok, Msg: msg}    // formateamos respuesta
	rJSON, err := json.Marshal(&r) // codificamos en JSON
	chk(err)                       // comprobamos error
	w.Write(rJSON)                 // escribimos el JSON resultante
}

func handler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()                              // es necesario parsear el formulario
	w.Header().Set("Content-Type", "text/plain") // cabecera estándar

	switch req.Form.Get("cmd") { // comprobamos comando desde el cliente
	case "hola": // ** registro
		response(w, true, "Hola "+req.Form.Get("mensaje"))
	default:
		response(w, false, "Comando inválido")
	}

}

func getMux() (mux *http.ServeMux) {
	mux = http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(handler))

	return
}

// Run : run sincronice server
func Run() {
	// suscripción SIGINT
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	mux := getMux()

	srv := &http.Server{Addr: ":8081", Handler: mux}

	// metodo concurrente
	go func() {
		if err := srv.ListenAndServeTLS("server.crt", "server.key"); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan // espera señal SIGINT
	log.Println("Apagando servidor ...")

	// apagar servidor de forma segura
	ctx, fnc := context.WithTimeout(context.Background(), 5*time.Second)
	fnc()
	srv.Shutdown(ctx)

	log.Println("Servidor detenido correctamente")
}
