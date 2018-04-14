package crypto

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/scrypt"
)

// función para comprobar errores (ahorra escritura)
func chk(e error) {
	if e != nil {
		panic(e)
	}
}

// Encode64 : función para codificar de []bytes a string(base64)
func Encode64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data) // sólo utiliza caracteres "imprimibles"
}

// Decode64 : función para decodificar de string(base64) a []bytes
func Decode64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s) // recupera el formato original
	chk(err)                                     // comprobamos el error
	return b                                     // devolvemos los datos originales
}

// Scrypt : genera salt y obtiene clave derivada
func Scrypt(pass []byte) (dk []byte, salt []byte) {
	salt = make([]byte, 32)
	_, err := rand.Read(salt)
	chk(err)

	dk, err = scrypt.Key(pass, salt, 1<<15, 8, 1, 32)
	chk(err)
	return
}

// ChkScrypt : comprueba que la contraseña sea correcta
func ChkScrypt(usrpass []byte, salt []byte, pass []byte) bool {
	newpass, err := scrypt.Key(pass, salt, 1<<15, 8, 1, 32)
	chk(err)
	return string(usrpass) == string(newpass)
}
