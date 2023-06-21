package protocol

import (
	"go.uber.org/zap"
	"net"
)

type Server struct {
	Addr string

	logger *zap.SugaredLogger
}

func (s *Server) Serve(l net.Listener) {

}
