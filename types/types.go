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

// User : tipo de usuario
type User struct {
	ID       string
	Username string
	Name     string
	Token    string
	Password string
}

// Folder : tipo de carpeta
type Folder struct {
	ID      string
	UserID  string
	Name    string
	Path    string
	Created string
	Updated string
	Folders []Folder
	Files   []File
}

// File : tipo de fichero
type File struct {
	ID       string
	FolderID string
}
