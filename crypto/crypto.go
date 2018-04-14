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
func Scrypt(pass []byte) (dk string, salt string) {
	bsalt := make([]byte, 32)
	_, err := rand.Read(bsalt)
	chk(err)

	key, err := scrypt.Key(pass, bsalt, 1<<15, 8, 1, 32)
	chk(err)
	dk = string(key)
	salt = string(bsalt)
	return
}

// ChkScrypt : comprueba que la contraseña sea correcta
func ChkScrypt(usrpass string, salt string, pass string) bool {
	newpass, err := scrypt.Key([]byte(pass), []byte(salt), 1<<15, 8, 1, 32)
	chk(err)
	return usrpass == string(newpass)
}
