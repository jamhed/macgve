package main

import (
	"awh/app"
	"fmt"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// Default is `-s -w -X main.version={{.Version}} -X main.commit={{.ShortCommit}} -X main.date={{.Date}} -X main.builtBy=goreleaser`.
var version string
var commit string
var date string
var builtBy string

var runtimeScheme = runtime.NewScheme()

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	_ = v1.AddToScheme(runtimeScheme)
}

func main() {
	fmt.Printf("Awh, version:%s commit:%s date:%s builtBy:%s\n", version, commit, date, builtBy)
	codecs := serializer.NewCodecFactory(runtimeScheme)
	deserializer := codecs.UniversalDeserializer()

	app := app.New(deserializer)
	app.Listen()
}
