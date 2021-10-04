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
