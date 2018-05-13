package types

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
	ID         int
	Email      string
	Name       string
	Token      string
	Password   []byte
	Salt       []byte
	MainFolder *Folder
}

// Folder : tipo de carpeta
type Folder struct {
	UserID  int
	Name    string
	Path    string
	Created string
	Updated string
	Folders []*Folder
	Files   []*File
}

// File : tipo de fichero
type File struct {
	ID       int
	FolderID string
}
