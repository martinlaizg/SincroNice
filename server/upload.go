package main

import (
	"SincroNice/types"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// loginHandler : manejador de la peticion a /login
func uploadHandler(w http.ResponseWriter, req *http.Request) {

	log.Println("File try upload")

	file, handler, err := req.FormFile("uploadfile")
	name := handler.Filename

	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	f, err := os.OpenFile("/home/pmga2/Escritorio/almacen/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	/*if _, exist := users[email]; exist {
		response(w, false, "ya existe un usuario con el mismo nombre de usuario")
		log.Println("Fail registry, user " + email + " already exist")
		return
	}*/
	NewFile := types.File{
		ID:       name,
		FolderID: "default"}

	files[name] = NewFile
	resp := types.Response{
		Status: true,
		Msg:    "Subido correctamente"}
	response(w, resp)
	log.Println("File " + name + " upload successful")
}
