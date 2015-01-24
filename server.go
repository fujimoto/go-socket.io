package socketio

import (
	"github.com/fujimoto/go-engine.io"
	"net/http"
	"time"
)

// Server is the server of socket.io.
type Server struct {
	*namespace
	broadcast BroadcastAdaptor
	eio       *engineio.Server
}

// NewServer returns the server supported given transports. If transports is nil, server will use ["polling", "websocket"] as default.
func NewServer(transportNames []string) (*Server, error) {
	eio, err := engineio.NewServer(transportNames)
	if err != nil {
		return nil, err
	}
	ret := &Server{
		namespace: newNamespace(newBroadcastDefault()),
		eio:       eio,
	}
	go ret.loop()
	return ret, nil
}

// SetPingTimeout sets the timeout of ping. When time out, server will close connection. Default is 60s.
func (s *Server) SetPingTimeout(t time.Duration) {
	s.eio.SetPingTimeout(t)
}

// SetPingInterval sets the interval of ping. Default is 25s.
func (s *Server) SetPingInterval(t time.Duration) {
	s.eio.SetPingInterval(t)
}

// SetMaxConnection sets the max connetion. Default is 1000.
func (s *Server) SetMaxConnection(n int) {
	s.eio.SetMaxConnection(n)
}

// SetAllowRequest sets the middleware function when establish connection. If it return non-nil, connection won't be established. Default will allow all request.
func (s *Server) SetAllowRequest(f func(*http.Request) error) {
	s.eio.SetAllowRequest(f)
}

// SetAllowUpgrades sets whether server allows transport upgrade. Default is true.
func (s *Server) SetAllowUpgrades(allow bool) {
	s.eio.SetAllowUpgrades(allow)
}

// SetAccessControl sets the middleware function which controls "Access-Control-Allow-*" header beaviors. Default will publish no access control headers
func (s *Server) SetAccessControl(f func(*http.Request) (orign string, credentials string, methods string, headers string, maxAge int)) {
	s.eio.SetAccessControl(f)
}

// SetCookie sets the name of cookie which used by engine.io. Default is "io".
func (s *Server) SetCookie(prefix string) {
	s.eio.SetCookie(prefix)
}

// SetAdaptor sets the adaptor of broadcast. Default is in-process broadcast implement.
func (s *Server) SetAdaptor(adaptor BroadcastAdaptor) {
	s.namespace = newNamespace(adaptor)
}

// ServeHTTP handles http request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.eio.ServeHTTP(w, r)
}

// Server level broadcasts function.
func (s *Server) BroadcastTo(room, message string, args ...interface{}) {
	s.namespace.BroadcastTo(room, message, args...)
}

func (s *Server) loop() {
	for {
		conn, err := s.eio.Accept()
		if err != nil {
			return
		}
		s := newSocket(conn, s.baseHandler)
		go func(s *socket) {
			s.loop()
		}(s)
	}
}
