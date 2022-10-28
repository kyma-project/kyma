package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"github.com/kyma-project/kyma/common/logging/logger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"telemetry-webhook-ca-init/internal"
)

const caName = "telemetry-validating-webhook-ca"

// const certDir = "/var/run/telemetry-webhook/"
var certDir string

func main() {
	flag.StringVar(&certDir, "cert-dir", "", "Path to certificate bundle directory")
	flag.Parse()

	// TODO debug
	certDir = "./bin"

	if err := validateFlags(); err != nil {
		panic(err.Error())
	}
	ctx := context.Background()
	log, err := logger.New("text", "info")
	if err != nil {
		panic(err.Error())
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	_, err = clientset.CoreV1().Secrets("kyma-system").Get(ctx, "test", v1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	caBundle, err := internal.CreateCABundle(caName)
	if err != nil {
		log.WithTracing(ctx).Error(err, "failed to create CA bundle")
		os.Exit(1)
	}

	err = os.MkdirAll(certDir, 0777)
	if err != nil {
		log.WithTracing(ctx).Error(err, "failed to create certs directory")
		os.Exit(1)
	}

	err = writeFile(certDir+"tls.crt", caBundle.ServerCert)
	if err != nil {
		log.WithTracing(ctx).Error(err, "failed to write tls.crt")
		os.Exit(1)
	}

	err = writeFile(certDir+"tls.key", caBundle.ServerPrivKey)
	if err != nil {
		log.WithTracing(ctx).Error(err, "failed to write tls.key")
		os.Exit(1)
	}

	// TODO get webhook config and set caBundle
}

func writeFile(filepath string, sCert *bytes.Buffer) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.Write(sCert.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func validateFlags() error {
	if certDir == "" {
		return errors.New("--cert-dir flag is required")
	}
	return nil
}
