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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/dgraph-io/badger/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func localBackup(storage *badger.DB, fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("fail to create file %s: %v", fileName, err)
	}
	defer f.Close()

	_, err = storage.Backup(f, 1)
	if err != nil {
		return fmt.Errorf("backup failed: %v", err)
	}

	fileInfo, err := f.Stat()
	if err != nil {
		badgerBackupSize.Set(-1)
		return fmt.Errorf("fail to get backup stats: %v", err)
	}
	badgerBackupSize.Set(float64(fileInfo.Size()))

	log.Print("local backup successfully created")

	return nil
}

type GoogleDriveInfo struct {
	dir    string
	token  *oauth2.Token
	config *oauth2.Config
}

func NewGoogleDriveInfo(dir string, token, credentials any) *GoogleDriveInfo {

	tb, _ := json.Marshal(token)
	t := &oauth2.Token{}
	json.Unmarshal(tb, t)

	b, _ := json.Marshal(credentials)
	config, _ := google.ConfigFromJSON(b, drive.DriveFileScope)

	return &GoogleDriveInfo{
		dir:    dir,
		token:  t,
		config: config,
	}
}

func (g *GoogleDriveInfo) saveBackupToGoogleDrive(fileName string) error {

	client := g.config.Client(context.Background(), g.token)

	service, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	dir, err := getDriveBackupDir(service, g.dir)
	if err != nil {
		return fmt.Errorf("unable to create dir: %v", err)
	}

	backup, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("fail to open backup file: %v", err)
	}
	defer backup.Close()

	file, err := uploadNewBackup(service, "backup.bak", "application/data", backup, dir.Id)
	if err != nil {
		return fmt.Errorf("fail to write backup in drive: %v", err)
	}

	log.Printf("Successfully uploaded %s to drive", file.Name)
	return nil
}

func getDriveBackupDir(service *drive.Service, dirName string) (*drive.File, error) {

	fileList, _ := service.Files.List().Do()
	for _, dir := range fileList.Files {
		if dir.Name == dirName {
			return dir, nil
		}
	}

	d := &drive.File{
		Name:     dirName,
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
