package app

import (
	"crypto/tls"
	"log"
	"macgve/args"
	"macgve/webhook"

	"k8s.io/apimachinery/pkg/runtime"
)

type App struct {
	args *args.Args
	srv  *webhook.Server
}

func New(deserializer runtime.Decoder) *App {
	a := new(App)
	a.args = args.New().LogLevel()
	pair, err := tls.LoadX509KeyPair(a.args.CertFile, a.args.KeyFile)
	if err != nil {
		log.Fatalf("Can't load cert and key files, cert:%s, key:%s, error:%s", a.args.CertFile, a.args.KeyFile, err)
	}
	a.srv = webhook.New(deserializer, a.args.Port, a.args.VaultAddr, a.args.GveImage, pair)
	return a
}

func (a *App) Listen() {
	a.srv.Listen()
}
