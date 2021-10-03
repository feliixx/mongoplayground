package internal

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var regIPV4 = regexp.MustCompile(`(\d+\.){3}(\d+):\d+`)

type LokiLogger struct {
	url      string
	nbToSend int

	httpClient *http.Client

	// pLock guards the payload bufffer
	pLock sync.Mutex
	// hold the message that will be send to loki
	payload *bytes.Buffer
}

func NewLokiLogger(host string, port int) *LokiLogger {

	l := &LokiLogger{
		url:     fmt.Sprintf("http://%s:%d/loki/api/v1/push", host, port),
		payload: bytes.NewBuffer(make([]byte, 0, 2048)),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 10 * time.Second,
				}).Dial,
				DisableKeepAlives: true, // no need for keep alive as we make only 1 req / 5 min
			},
		},
	}
	l.reset()
	return l
}

func (l *LokiLogger) Write(msg []byte) (int, error) {

	l.nbToSend++
	// Print the message to stdout just in case there is a problem
	// with the loki server
	fmt.Printf("%s", msg)

	l.pLock.Lock()
	defer l.pLock.Unlock()

	// loki requires that logs doesn't end with trailing newline
	msg = bytes.TrimRight(msg, "\n")
	// anonymise any IP adress
	msg = regIPV4.ReplaceAll(msg, []byte("x.x.x.x"))

	l.payload.WriteString(`["`)
	l.payload.WriteString(strconv.Itoa(int(time.Now().Unix())))
	l.payload.WriteString(`000000000","`)
	// if the message starts with a datetime like 2021/10/12 10:11:00,
	// remove it
	if bytes.HasPrefix(msg, []byte("202")) {
		msg = msg[20:]
	}
	l.payload.Write(msg)
	l.payload.WriteString(`"],`)

	return len(msg), nil
}

// Send send the logs in the buffer to loki.
// The message sent looks like this:
//
// {
//   "streams": [
//     {
//       "stream": {
//         "app": "mongoplayground"
//       },
//       "values": [
//         [
//           "1633228233000000000",
//           "fail to load page with id v : invalid page id length"
//         ],
//         [
//           "1633228235000000000",
//           "fail to load page with id vk : invalid page id length"
//         ]
//       ]
//     }
//   ]
// }
func (l *LokiLogger) Send() error {

	if l.nbToSend == 0 {
		return nil
	}

	l.pLock.Lock()

	// remove the extra "," added by the last call to `Write`
	l.payload.Truncate(l.payload.Len() - 1)
	l.payload.WriteString(`]}]}`)

	// make sure that the duration of the POST request can't affect the response
	// to other request by grabbing a copy of the message in order to release the
	// lock before sending the POST request
	body := l.payload.Bytes()
	l.reset()
	l.pLock.Unlock()

	resp, err := l.httpClient.Post(l.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	io.Copy(io.Discard, resp.Body)
	return resp.Body.Close()
}

func (l *LokiLogger) reset() {
	l.payload.Reset()
	l.payload.WriteString(`{"streams": [{ "stream": { "app": "mongoplayground" }, "values": [ `)
	l.nbToSend = 0
}
