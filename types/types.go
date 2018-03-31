package types

type Resp struct {
	Status bool
	Msg    string
}

type Heredado struct {
	Resp
	token string
}
