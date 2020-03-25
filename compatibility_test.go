package main

import (
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

						err = testServer.session.Database(p.dbHash()).Drop(nil)
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
