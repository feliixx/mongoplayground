// mongoplayground: a sandbox to test and share MongoDB queries
// Copyright (C) 2017 Adrien Petel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/badger/v2"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	// google drive dir for storing last backup
	driveBackupDir = "autobackup"
	// file holding google drive token
	tokenFile = "token.json"
)

// create a backup from the badger db, and store it in backupDir.
// keep a backup of last seven days only. Older backups are
// overwritten
// upload the last backup to google drive. Previous backup is moved to trash
// and automatically removed after 30 days
func (s *Server) backup() {

	if _, err := os.Stat(s.backupDir); os.IsNotExist(err) {
		os.Mkdir(s.backupDir, os.ModePerm)
	}

	fileName := fmt.Sprintf("%s/badger_%d.bak", s.backupDir, time.Now().Weekday())

	localBackup(s.logger, s.storage, fileName)
	saveBackupToGoogleDrive(s.logger, fileName)
}

func localBackup(log *log.Logger, storage *badger.DB, fileName string) {
	f, err := os.Create(fileName)
	if err != nil {
		log.Printf("fail to create file %s: %v", fileName, err)
	}
	defer f.Close()

	_, err = storage.Backup(f, 1)
	if err != nil {
		log.Printf("backup failed: %v", err)
	}

	fileInfo, err := f.Stat()
	if err != nil {
		log.Printf("fail to get backup stats")
	}
	badgerBackup.Set(float64(fileInfo.Size()))
}

func saveBackupToGoogleDrive(log *log.Logger, fileName string) {

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

	dir, err := getDriveBackupDir(service)
	if err != nil {
		log.Printf("Unable to create dir: %v", err)
		return
	}

	backup, err := os.Open(fileName)
	if err != nil {
		log.Printf("Fail to open backup file: %v", err)
		return
	}
	defer backup.Close()

	file, err := uploadNewBackup(service, "backup.bak", "application/data", backup, dir.Id)
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

func getDriveBackupDir(service *drive.Service) (*drive.File, error) {

	fileList, _ := service.Files.List().Do()
	for _, dir := range fileList.Files {
		if dir.Name == driveBackupDir {
			return dir, nil
		}
	}

	d := &drive.File{
		Name:     driveBackupDir,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{"root"},
	}
	return service.Files.Create(d).Do()
}

func uploadNewBackup(service *drive.Service, name string, mimeType string, content io.Reader, parentID string) (*drive.File, error) {

	fileList, _ := service.Files.List().Do()
	for _, f := range fileList.Files {
		if f.Name == name {
			// previous backup is moved to trash.
			// files in trash for more than 30 days are automatically removed
			service.Files.Update(f.Id, &drive.File{Trashed: true}).Do()
		}
	}

	f := &drive.File{
		MimeType: mimeType,
		Name:     name,
		Parents:  []string{parentID},
	}
	return service.Files.Create(f).Media(content).Do()
}
