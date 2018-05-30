package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
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

// Hash : genera el hash de data en [64]byte
func Hash(data []byte) [64]byte {
	return sha512.Sum512(data)
}

// Encrypt : función para cifrar (con AES en este caso), adjunta el IV al principio
func Encrypt(data, key []byte) (out []byte) {
	out = make([]byte, len(data)+16)    // reservamos espacio para el IV al principio
	rand.Read(out[:16])                 // generamos el IV
	blk, err := aes.NewCipher(key)      // cifrador en bloque (AES), usa key
	chk(err)                            // comprobamos el error
	ctr := cipher.NewCTR(blk, out[:16]) // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out[16:], data)    // ciframos los datos
	return
}

// Decrypt : función para descifrar (con AES en este caso)
func Decrypt(data, key []byte) (out []byte) {
	out = make([]byte, len(data)-16)     // la salida no va a tener el IV
	blk, err := aes.NewCipher(key)       // cifrador en bloque (AES), usa key
	chk(err)                             // comprobamos el error
	ctr := cipher.NewCTR(blk, data[:16]) // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out, data[16:])     // desciframos (doble cifrado) los datos
	return
}
