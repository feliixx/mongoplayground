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
	"io"
	"log"
	"net/http"
	"regexp"
	"testing"
)

const (
	lokiPort = 65000
)

var (
	reqBody      = []byte{}
	regTimestamp = regexp.MustCompile(`(\d{15}0000)`)
)

func TestSendLogsToLoki(t *testing.T) {

	startMockLokiServer()

	l := NewLokiLogger("localhost", lokiPort)

	l.Write([]byte("first log message"))
	l.Write([]byte("second log message with line return\n\n"))
	l.Write([]byte("third log message with an IP: 172.0.0.1:65112"))
	l.Write([]byte(`2021/10/03 16:31:42 goroutine 28 [running]:
runtime/debug.Stack()
	/usr/local/go/src/runtime/debug/stack.go:24 +0x65
github.com/feliixx/mongoplayground/internal.latencyObserver.func1.1()
	/home/adrien/project/mongoplayground/internal/server.go:86 +0x2a
panic({0xe3dba0, 0xc0004b4210})`))

	err := l.Send()
	if err != nil {
		t.Errorf("fail to send logs: %v", err)
	}

	want := `{"streams": [{ "stream": { "app": "mongoplayground" }, "values": [ ["0000000000000000000","first log message"],["0000000000000000000","second log message with line return"],["0000000000000000000","third log message with an IP: x.x.x.x"],["0000000000000000000","goroutine 28 [running]:\nruntime/debug.Stack()\n\t/usr/local/go/src/runtime/debug/stack.go:24 +0x65\ngithub.com/feliixx/mongoplayground/internal.latencyObserver.func1.1()\n\t/home/adrien/project/mongoplayground/internal/server.go:86 +0x2a\npanic({0xe3dba0, 0xc0004b4210})"]]}]}`
	got := string(regTimestamp.ReplaceAll(reqBody, []byte("0000000000000000000")))

	if want != got {
		t.Errorf("Got wrong body:\n expected:\n\n%v\n\n but got\n\n%v\n", want, got)
	}

}

func startMockLokiServer() {

	http.HandleFunc("/loki/api/v1/push", func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ = io.ReadAll(r.Body)
	})
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", lokiPort), nil))
	}()
}
