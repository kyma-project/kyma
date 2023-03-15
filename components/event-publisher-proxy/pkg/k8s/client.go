package k8s

import (
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func ClientOrDie(cfg *rest.Config) kubernetes.Interface {
	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("error:[%v]", err)
	}
	return c
}
