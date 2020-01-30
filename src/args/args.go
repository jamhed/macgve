package args

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
)

type Args struct {
	VerboseLevel string
	Port         int
	CertFile     string
	KeyFile      string
	Args         []string
}

func New() *Args {
	return new(Args).Parse()
}

func env(name, def string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return def
}

func (a *Args) Parse() *Args {
	flag.StringVar(&a.VerboseLevel, "verbose", env("VERBOSE", "info"), "Set verbosity level")
	flag.StringVar(&a.CertFile, "certFile", env("CERT_FILE", "cert.pem"), "Cert file path")
	flag.StringVar(&a.KeyFile, "keyFile", env("KEY_FILE", "key.pem"), "Key file path")
	flag.IntVar(&a.Port, "port", 8443, "Listen port")
	flag.Parse()
	a.Args = flag.Args()
	return a
}

func (a *Args) LogLevel() *Args {
	switch a.VerboseLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
	return a
}
