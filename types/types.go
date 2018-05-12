package types

import (
	"github.com/rs/xid"
)

// Response : base de la respuesta al cliente
type Response struct {
	Status bool
	Msg    string
}

// Heredado : estructura de respuesta heredada (es un ejemplo)
type Heredado struct {
	Response
	token string
}

// User : tipo de usuario
type User struct {
	ID         string
	Email      string
	Name       string
	Token      string
	Password   []byte
	Salt       []byte
	MainFolder string
}

// Folder : tipo de carpeta
type Folder struct {
	ID      string
	UserID  string
	Name    string
	Path    string
	Created string
	Updated string
	Folders map[string]string
	Files   map[string]string
}

// File : tipo de fichero
type File struct {
	ID       int
	FolderID string
}

func GenXid() string {
	id := xid.New()
	generated := id.String()
	return generated
}
