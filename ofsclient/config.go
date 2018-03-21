package ofsclient

// Config defines configuration information for an ofs client
type Config struct {
	// A collection of servers used to allocate flakes
	Servers []ServerEntry `yaml:"servers"`
}

// ServerEntry defines information about an ofs server
type ServerEntry struct {
	// Server is the full address (host:port) of the ofs server
	Server string `yaml:"server"`
	// Auth is the auth code for the server (<255 bytes), or "" for no auth
	Auth string `yaml:"auth"`
}
