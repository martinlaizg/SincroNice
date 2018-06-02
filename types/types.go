package types

import (
	"github.com/rs/xid"
)

// Response : base de la respuesta al cliente
type Response struct {
	Status bool
	Msg    string
}

// ResponseToken : respuesta con el token
type ResponseToken struct {
	Response
	Token string
}

// ResponseLogin : respuesta al login
type ResponseLogin struct {
	Response
	User
}

// User : tipo de usuario
type User struct {
	ID         string
	Email      string
	Name       string
	Token      string
	FilePass   string
	Password   []byte
	Salt       []byte
	MainFolder string
}

// Folder : tipo de carpeta
type Folder struct {
	ID           string
	UserID       string
	Name         string
	Path         string
	Created      string
	Updated      string
	FolderParent string
	Folders      map[string]string
	Files        map[string]string
}

// File : tipo de fichero
type File struct {
	ID       string
	Name     string
	FolderID string
	OwnerID  string
	Versions []Version
}

// Version : versión
type Version struct {
	ID     string
	Blocks []string
}

// Block : tipo de bloque
type Block struct {
	ID    string
	Hash  []byte
	Owner string
}

// GenXid : generador de bloques
func GenXid() string {
	id := xid.New()
	generated := id.String()
	return generated
}
