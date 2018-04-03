package main

import (
	"SincroNice/types"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/howeyc/gopass"
)

var baseURL = "https://localhost:8081"

var client *http.Client

func chk(e error) {
	if e != nil {
		panic(e)
	}
}

// send : enva {data} a la url localhost:8081/{endpoint}
func send(endpoint string, data url.Values) *http.Response {
	r, err := client.PostForm(baseURL+endpoint, data)
	chk(err)
	return r
}

func menu() {
	fmt.Println("Bienvenido a SincroNice")
	fmt.Println("Login")
	fmt.Printf("Username: ")
	var usr string
	fmt.Scanf("%s\n", &usr)
	fmt.Printf("Pass: ")
	pass, err := gopass.GetPasswdMasked()
	chk(err)

	log.Println("Login as " + usr)

	data := url.Values{}
	data.Set("usr", usr)
	data.Set("pass", string(pass))

	response := send("/login", data)
	bData, err := ioutil.ReadAll(response.Body)
	chk(err)
	var rData types.Resp
	err = json.Unmarshal(bData, &rData)
	chk(err)

	if rData.Status == true {
		fmt.Printf("Logeado correctamente\n")
		return
	}
	fmt.Printf("Error al loguear: %v\n", rData.Msg)

}

func createClient() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
}

// RunClient : run sincronice client
func main() {
	createClient()
	menu()
}
