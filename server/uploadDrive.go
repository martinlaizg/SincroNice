package main

import (
	"SincroNice/crypto"
	"SincroNice/types"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/antonholmquist/jason"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

//Ruta donde guardamos los archivos en el servidor
var ruta = "C:/Users/pedro/go/src/SincroNice/server/tmp/"

func checkBlockHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")
	hash := crypto.Decode64(req.Form.Get("hash"))
	r.Status = false
	r.Msg = types.GenXid()
	for _, block := range blocks {
		if string(block.Hash) == string(hash) {
			r.Status = true
			r.Msg = block.ID
		}
	}
	response(w, r)
}

func uploadDriveHandler(w http.ResponseWriter, req *http.Request) {

	req.ParseMultipartForm(1)
	r := types.Response{}
	w.Header().Set("Content-Type", "application/json")
	blockID := string(crypto.Decode64(req.PostFormValue("blockID")))
	userID := string(crypto.Decode64(req.PostFormValue("userID")))
	block, _, err := req.FormFile("fileupload") // Obtenemos el fichero
	defer block.Close()
	chk(err)
	fmt.Printf(blockID + "/" + userID)

	blockBytes, err := ioutil.ReadAll(block) // Lo pasamos a bytes
	chk(err)

	hash := crypto.Hash(blockBytes)

	blockT := types.Block{
		ID:    blockID,
		Hash:  hash[:],
		Owner: userID,
	}
	blocks[blockT.ID] = blockT

	newPath := ruta + blockID
	newBlock, err := os.Create(newPath)
	defer newBlock.Close()
	chk(err)
	_, err = newBlock.Write(blockBytes)
	newBlock.Sync()
	chk(err)
	r.Status = true
	response(w, r)

	//upload drive
	ctx := context.Background()
	// process the credential file
	credential, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(credential, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(ctx, config)

	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}

	token, err := tokenFromFile(cacheFile)
	if err != nil {
		log.Fatalf("Unable to get token from file. %v", err)
	}

	id := "1SXfqr0Jm6iEe04W5BGvo2X57pYvatDjY"
	DownloadFile(client, id)

	fileBytes, err := ioutil.ReadFile(ruta + blockID)
	if err != nil {
		log.Fatalf("Unable to read file for upload: %v", err)
	}

	fileMIMEType := http.DetectContentType(fileBytes)

	postURL := "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart"
	authToken := token.AccessToken

	boundary := randStr(32, "alphanum")

	uploadData := []byte("\n" +
		"--" + boundary + "\n" +
		"Content-Type: application/json; charset=" + string('"') + "UTF-8" + string('"') + "\n\n" +
		"{ \n" +
		string('"') + "name" + string('"') + ":" + string('"') + blockID + string('"') + "\n" +
		"} \n\n" +
		"--" + boundary + "\n" +
		"Content-Type:" + fileMIMEType + "\n\n" +
		string(fileBytes) + "\n" +

		"--" + boundary + "--")

	// post to Drive with RESTful method
	request, _ := http.NewRequest("POST", postURL, strings.NewReader(string(uploadData)))
	request.Header.Add("Host", "www.googleapis.com")
	request.Header.Add("Authorization", "Bearer "+authToken)
	request.Header.Add("Content-Type", "multipart/related; boundary="+string('"')+boundary+string('"'))
	request.Header.Add("Content-Length", strconv.FormatInt(request.ContentLength, 10))

	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("Unable to be post to Google API: %v", err)
		return
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatalf("Unable to read Google API response: %v", err)
		return
	}

	fmt.Println(string(body))

	log.Println("File " + blockID + " upload successful")

}

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("google-drive-golang.json")), err
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func randStr(strSize int, randType string) string {

	var dictionary string

	if randType == "alphanum" {
		dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "alpha" {
		dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "number" {
		dictionary = "0123456789"
	}

	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}

/*
func uploadDriveHandler(w http.ResponseWriter, req *http.Request) {

	log.Println("File try upload")
	fmt.Println("File try upload")
	fil, handler, err := req.FormFile("uploadfile")
	fileName := handler.Filename
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fil.Close()

	f, err := os.OpenFile(ruta+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, fil)

	ctx := context.Background()
	// process the credential file
	credential, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(credential, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(ctx, config)

	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}

	token, err := tokenFromFile(cacheFile)
	if err != nil {
		log.Fatalf("Unable to get token from file. %v", err)
	}

	id := "1SXfqr0Jm6iEe04W5BGvo2X57pYvatDjY"
	DownloadFile(client, id)

	fileBytes, err := ioutil.ReadFile(ruta + fileName)
	if err != nil {
		log.Fatalf("Unable to read file for upload: %v", err)
	}

	fileMIMEType := http.DetectContentType(fileBytes)

	postURL := "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart"
	authToken := token.AccessToken

	boundary := randStr(32, "alphanum")

	uploadData := []byte("\n" +
		"--" + boundary + "\n" +
		"Content-Type: application/json; charset=" + string('"') + "UTF-8" + string('"') + "\n\n" +
		"{ \n" +
		string('"') + "name" + string('"') + ":" + string('"') + fileName + string('"') + "\n" +
		"} \n\n" +
		"--" + boundary + "\n" +
		"Content-Type:" + fileMIMEType + "\n\n" +
		string(fileBytes) + "\n" +

		"--" + boundary + "--")

	// post to Drive with RESTful method
	request, _ := http.NewRequest("POST", postURL, strings.NewReader(string(uploadData)))
	request.Header.Add("Host", "www.googleapis.com")
	request.Header.Add("Authorization", "Bearer "+authToken)
	request.Header.Add("Content-Type", "multipart/related; boundary="+string('"')+boundary+string('"'))
	request.Header.Add("Content-Length", strconv.FormatInt(request.ContentLength, 10))

	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("Unable to be post to Google API: %v", err)
		return
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatalf("Unable to read Google API response: %v", err)
		return
	}

	fmt.Println(string(body))

	log.Println("File " + fileName + " upload successful")
}
*/
// DownloadFile downloads the content of a given file object
func DownloadFile(client *http.Client, id string) (string, error) {
	// t parameter should use an oauth.Transport

	downloadURL := "https://www.googleapis.com/drive/v2/files/" + id
	if downloadURL == "" {
		// If there is no downloadUrl, there is no body
		fmt.Printf("An error occurred: File is not downloadable")
		return "", nil
	}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return "", err
	}

	response, err := client.Do(req)

	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return "", err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return "", err
	}

	fmt.Printf(string(body))
	fmt.Printf("///////////////////")

	json, _ := jason.NewObjectFromBytes(body)
	url, _ := json.GetString("webContentLink")
	fmt.Println("url : ", url)
	fmt.Printf("///////////////////")

	//fileMIMEType := http.DetectContentType(fileBytes)

	return string(body), nil
}
