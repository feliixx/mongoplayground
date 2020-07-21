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

package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v2"
)

const backupPath = "backups/backup.bak"

func TestGenerateresultFile(t *testing.T) {

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Skip("backup file doesn't exist")
	}

	testServer.clearDatabases(t)

	backup, err := os.Open(backupPath)
	if err != nil {
		t.Errorf("fail to open backup file: %v", err)
	}

	testServer.storage.Load(backup, 0)

	out, err := os.Create("backups/new_result.txt")
	if err != nil {
		t.Errorf("fail to create result file: %v", err)
	}
	defer out.Close()

	err = testServer.storage.View(func(txn *badger.Txn) error {

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			key := item.Key()
			if string(key) != "6wGof2r4ZWH" {

				item.Value(func(val []byte) error {

					p := &page{}
					p.decode(val)

					result, err := testServer.run(p)
					if err == nil {

						fmt.Println(string(key))

						out.Write(key)
						out.WriteString(":")
						out.Write(result)
						out.WriteString("\n")

						err = testServer.session.Database(p.dbHash()).Drop(context.Background())
						if err != nil {
							fmt.Printf("fail to drop db: %v", err)
						}
						delete(testServer.activeDB, p.dbHash())
					}
					return nil
				})
			}
		}
		return nil
	})
	if err != nil {
		t.Errorf("fail to get pages: %v", err)
	}
}
