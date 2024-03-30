package ArgsParser

import "flag"

// ArgsParser struct
type ArgsParser struct {
	Ip       *string
	Port     *int
	Secure   *bool
	CertFile *string
	KeyFile  *string
}

// Ap is a global ArgsParser
var Ap *ArgsParser

func init() {
	// Create a new ArgsParser
	Ap = &ArgsParser{}

	// Get the flags
	Ap.Ip = flag.String("ip", "0.0.0.0", "The IP to bind the server to")
	Ap.Port = flag.Int("port", 8080, "The port to bind the server to")
	Ap.Secure = flag.Bool("secure", false, "Use HTTPS")
	Ap.CertFile = flag.String("certfile", "", "The path to the certificate file")
	Ap.KeyFile = flag.String("keyfile", "", "The path to the key file")
}
