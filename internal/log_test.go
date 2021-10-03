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
	l.Write([]byte("second log message"))
	l.Write([]byte("third log message"))

	err := l.Send()
	if err != nil {
		t.Errorf("fail to send logs: %v", err)
	}

	want := `{"streams": [{ "stream": { "app": "mongoplayground" }, "values": [ ["0000000000000000000","first log message"],["0000000000000000000","second log message"],["0000000000000000000","third log message"]]}]}`
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
