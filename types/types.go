package types

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

// User :
type User struct {
	id       string
	username string
	name     string
	token    string
	password string
}

// Folder :
type Folder struct {
	ID      string
	Name    string
	Path    string
	Created string
	Updated string
	Folders []Folder
	Files   []File
}

// File :
type File struct {
	id       string
	folderID string
}
