package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultNamespace = "default"

	defaultValidityTime = 184 * 24 * time.Hour
)

type options struct {
	connectorCertificateSecret types.NamespacedName
	caCertificateSecret        types.NamespacedName

	caCertificate string
	caKey         string

	caCertificateSecretToMigrate     types.NamespacedName
	caCertificateSecretKeysToMigrate string

	connectorCertificateSecretToMigrate     types.NamespacedName
	connectorCertificateSecretKeysToMigrate string

	generatedValidityTime time.Duration
}

func parseArgs() *options {
	connectorCertificateSecret := flag.String("connectorCertificateSecret", "kyma-integration/connector-service-app-ca", "Secret namespace/name used by the Connector Service")
	caCertificateSecret := flag.String("caCertificateSecret", "istio-system/kyma-gateway-certs-cacert", "Secret namespace/name where CA certificate is kept")

	caCertificate := flag.String("caCertificate", "", "Base64 encoded pem CA certificate")
	caKey := flag.String("caKey", "", "Base64 encoded pem CA key")

	caCertificateSecretToMigrate := flag.String("caCertificateSecretToMigrate", "istio-system/app-connector-certs", "Name of the secret containing CA root to be migrated. Use when there is a need to rename a secret.")
	caCertificateSecretKeysToMigrate := flag.String("caCertificateSecretKeysToMigrate", `["cacert"]`, "List of keys from certificate secret to be migrated")

	connectorCertificateSecretToMigrate := flag.String("connectorCertificateSecretToMigrate", "", "Name of the secret containing CA root to be migrated. Use when there is a need to rename a secret.")
	connectorCertificateSecretKeysToMigrate := flag.String("connectorCertificateSecretKeysToMigrate", `["ca.crt", "ca.key", "cacert"]`, "List of keys from certificate secret to be migrated")

	generatedValidityTime := flag.String("generatedValidityTime", "30d", "Validity time of the generated certificate")

	flag.Parse()

	validityTime, err := parseDuration(*generatedValidityTime)
	if err != nil {
		logrus.Infof("Failed to parse validity time for generated certificate: %s, using default value %d.", err, defaultValidityTime)
	}

	return &options{
		connectorCertificateSecret:              parseNamespacedName(*connectorCertificateSecret),
		caCertificateSecret:                     parseNamespacedName(*caCertificateSecret),
		caCertificate:                           *caCertificate,
		caKey:                                   *caKey,
		caCertificateSecretToMigrate:            parseNamespacedName(*caCertificateSecretToMigrate),
		caCertificateSecretKeysToMigrate:        *caCertificateSecretKeysToMigrate,
		connectorCertificateSecretToMigrate:     parseNamespacedName(*connectorCertificateSecretToMigrate),
		connectorCertificateSecretKeysToMigrate: *connectorCertificateSecretKeysToMigrate,
		generatedValidityTime:                   validityTime,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--connectorCertificateSecret=%s --caCertificateSecret=%s "+
		"-caCertificateSecretToMigrate=%s --caCertificateSecretKeysToMigrate=%s"+
		"--generatedValidityTime=%s "+
		"CA certificate provided: %t, CA key provided: %t",
		o.connectorCertificateSecret, o.caCertificateSecret,
		o.caCertificateSecretToMigrate,
		o.caCertificateSecretKeysToMigrate,
		o.generatedValidityTime.String(),
		o.caCertificate != "", o.caKey != "")
}

func parseDuration(durationString string) (time.Duration, error) {
	unitsMap := map[string]time.Duration{"m": time.Minute, "h": time.Hour, "d": 24 * time.Hour}

	timeUnit := durationString[len(durationString)-1:]
	_, ok := unitsMap[timeUnit]
	if !ok {
		return defaultValidityTime, fmt.Errorf("unrecognized time unit provided: %s", timeUnit)
	}

	timeLength, err := strconv.Atoi(durationString[:len(durationString)-1])
	if err != nil {
		return defaultValidityTime, err
	}

	return time.Duration(timeLength) * unitsMap[timeUnit], nil
}

func parseNamespacedName(value string) types.NamespacedName {
	parts := strings.Split(value, string(types.Separator))

	if singleValueProvided(parts) {
		return types.NamespacedName{
			Name:      parts[0],
			Namespace: defaultNamespace,
		}
	}

	namespace := get(parts, 0)
	if namespace == "" {
		namespace = defaultNamespace
	}

	return types.NamespacedName{
		Namespace: namespace,
		Name:      get(parts, 1),
	}
}

func singleValueProvided(split []string) bool {
	return len(split) == 1 || get(split, 1) == ""
}

func get(array []string, index int) string {
	if len(array) > index {
		return array[index]
	}
	return ""
}
