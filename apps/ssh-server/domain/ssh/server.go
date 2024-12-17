package ssh

type Server struct {
	Port       int
	PrivateKey string
}

func (s *Server) Serve() error {
	return nil
}

func (s *Server) Shutdown() error {
	return nil
}
