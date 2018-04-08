package types

type Resp struct {
	Status bool
	Msg    string
}

type Heredado struct {
	Resp
	token string
}

// User :
type User struct {
	id       string
	username string
	name     string
	token    string
}

// Folder :
type Folder struct {
	id      string
	name    string
	path    string
	folders []Folder
	files   []File
}

// File :
type File struct {
	id       string
	folderID string
}
