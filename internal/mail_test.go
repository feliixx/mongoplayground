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
	"net/http"
	"net/url"
	"testing"
)

func TestPrettyPrintRequest(t *testing.T) {

	r, _ := http.NewRequest("POST", "/run", nil)
	r.Header.Set("Content-encoding", "br")
	r.PostForm = url.Values{
		"mode": {"bson"},
	}

	want := "request: /run\n\n[Content-Encoding]: br\n\nbody\n[mode]: bson\n"
	if got := prettyPrintRequest(r); want != got {
		t.Errorf("Text doesn't match: expected\n%s\nbut got\n%s", want, got)
	}
}

func TestCreateMessage(t *testing.T) {

	want := "Subject: [Mongoplayground] Panic\r\n\r\nsome error msg\r\n"
	got := string(createMessage("Panic", "some error msg"))
	if want != got {
		t.Errorf("Text doesn't match: expected\n%s\nbut got\n%s", want, got)
	}

}
