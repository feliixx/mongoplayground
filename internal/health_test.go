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
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestHealthCheck(t *testing.T) {

	buf := httpBody(t, testServer.healthHandler, http.MethodGet, healthEndpoint, url.Values{})

	want := fmt.Sprintf(`{"Status":"UP","Services":[{"Name":"badger","Status":"UP"},{"Name":"mongodb","Version":"%s","Status":"UP"}],"Version":""}`, testServer.mongodbVersion)
	if got := buf.String(); want != got {
		t.Errorf("expected\n%s\nbut got\n%s", want, got)
	}
}
