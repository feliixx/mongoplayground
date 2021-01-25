package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const tokenFile = "token.json"

func saveBackupToGoogleDrive(log *log.Logger, backup io.Reader) {

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Printf("Unable to read client secret file: %v", err)
		return
	}

	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		log.Printf("Unable to parse client secret file to config: %v", err)
		return
	}
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		log.Printf("Fail to read token file: %v", err)
		return
	}
	client := config.Client(context.Background(), token)

	service, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Unable to retrieve Drive client: %v", err)
		return
	}

	dir, err := createDir(service, "autobackup", "root")
	if err != nil {
		log.Printf("Unable to create dir: %v", err)
		return
	}

	file, err := createFile(service, "backup.bak", "", backup, dir.Id)
	if err != nil {
		log.Printf("Fail to write backup in drive: %v", err)
		return
	}
	log.Printf("Successfully uploaded %s to drive", file.Name)
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func createDir(service *drive.Service, name string, parentID string) (*drive.File, error) {
	d := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}
	return service.Files.Create(d).Do()
}

func createFile(service *drive.Service, name string, mimeType string, content io.Reader, parentID string) (*drive.File, error) {
	f := &drive.File{
		MimeType: mimeType,
		Name:     name,
		Parents:  []string{parentID},
	}
	return service.Files.Create(f).Media(content).Do()
}
