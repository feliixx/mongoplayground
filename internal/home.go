package internal

import "net/http"

// return a playground with the default configuration
func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Encoding", "gzip")

	w.Write(s.staticContent[homeEndpoint])
}
