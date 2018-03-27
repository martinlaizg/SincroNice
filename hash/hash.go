package main

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// función para comprobar errores (ahorra escritura)
func chk(e error) {
	if e != nil {
		panic(e)
	}
}

// función para codificar de []bytes a string (Base64)
func encode64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data) // sólo utiliza caracteres "imprimibles"
}

// función para decodificar de string a []bytes (Base64)
func decode64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s) // recupera el formato original
	chk(err)                                     // comprobamos el error
	return b                                     // devolvemos los datos originales
}

func main() {

	// creamos usuario
	var login, password string
	fmt.Print("Introduce tu login: ")
	fmt.Scanf("%s\n", &login)

	// contraseña
	fmt.Print("Introduce tu password: ")
	fmt.Scanf("%s\n", &password)

	// codificación base64
	passBase64 := encode64([]byte(password))
	fmt.Printf("Password base64: %s \n", passBase64)

	// decoficiación base64
	pass := decode64(passBase64)
	fmt.Printf("Password orig: %s \n", pass)

	// hash con SHA512 de la contraseña
	passwordHash := sha512.Sum512([]byte(password))
	slice := passwordHash[:]

	// codificación base64
	hashBase64 := encode64(slice)
	fmt.Printf("Password base64(sha512(password)): %s \n", hashBase64)

	// sha512 + salt
	hashedPassword, err := bcrypt.GenerateFromPassword(slice, bcrypt.DefaultCost)

	if err != nil {
		panic(err)
	}

	fmt.Println(string(hashedPassword))

	err = bcrypt.CompareHashAndPassword(hashedPassword, slice)

	fmt.Println(err)

}
