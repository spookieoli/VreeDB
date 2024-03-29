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

// NewArgsParser returns a new ArgsParser // TODO: SUBJECT TO HEAVY CHANGE -- use flags instead
func NewArgsParser() *ArgsParser {
	// Create an new ArgsParser
	ap := &ArgsParser{}

	// Get the flags
	ap.Ip = flag.String("ip", "0.0.0.0", "The IP to bind the server to")
	ap.Port = flag.Int("port", 8080, "The port to bind the server to")
	ap.Secure = flag.Bool("secure", false, "Use HTTPS")
	ap.CertFile = flag.String("certfile", "", "The path to the certificate file")
	ap.KeyFile = flag.String("keyfile", "", "The path to the key file")
	return ap
}
