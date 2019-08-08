package main

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	http.Server
	connCount    int32
	shuttingDown bool
	sync.Mutex
}

func Wrap(hs http.Server) *Server {
	s := &Server{Server: hs}
	s.Handler = s.wrapHandler(s.Handler)
	return s
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.Lock()
	s.shuttingDown = true
	s.Unlock()

	for {
		s.Lock()
		if s.connCount == 0 {
			s.Unlock()
			break
		}
		s.Unlock()

		time.Sleep(time.Millisecond)
	}

	return s.Server.Shutdown(ctx)
}

func (s *Server) wrapHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Lock()
		if s.shuttingDown {
			w.WriteHeader(503)
			io.WriteString(w, "server shutting down")
			s.Unlock()
			return
		}
		s.connCount++
		s.Unlock()

		handler.ServeHTTP(w, r)

		s.Lock()
		s.connCount--
		s.Unlock()
	})
}
