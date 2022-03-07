package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"macgve/mutate"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Server struct {
	vaultaddr    string
	gveimage     string
	deserializer runtime.Decoder
	server       *http.Server
}

type serverErrorLogWriter struct{}

func (*serverErrorLogWriter) Write(p []byte) (int, error) {
	m := string(p)
	logrus.Error(m)
	if strings.HasPrefix(m, "http: TLS handshake error") {
		os.Exit(1)
	}
	return len(p), nil
}

func New(deserializer runtime.Decoder, port int, vaultaddr, gveimage string, pair tls.Certificate) *Server {
	srv := &Server{
		deserializer: deserializer,
		server: &http.Server{
			ErrorLog:  log.New(&serverErrorLogWriter{}, "", 0),
			Addr:      fmt.Sprintf(":%v", port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/pods", srv.Serve)
	srv.server.Handler = mux
	srv.vaultaddr = vaultaddr
	srv.gveimage = gveimage
	return srv
}

func (srv *Server) Listen() {
	go func() {
		if err := srv.server.ListenAndServeTLS("", ""); err != nil {
			logrus.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	logrus.Info("Server started")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	logrus.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	srv.server.Shutdown(context.Background())
}

func (srv *Server) Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		logrus.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logrus.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1.AdmissionResponse
	ar := v1.AdmissionReview{}
	if _, _, err := srv.deserializer.Decode(body, nil, &ar); err != nil {
		logrus.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
	} else {
		logrus.Infof("request: %s", r.URL.Path)
		if r.URL.Path == "/pods" {
			admissionResponse = mutate.Mutate(&ar, srv.vaultaddr, srv.gveimage)
		}
	}

	admissionReview := v1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "AdmissionReview",
		},
	}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		logrus.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	if _, err := w.Write(resp); err != nil {
		logrus.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
