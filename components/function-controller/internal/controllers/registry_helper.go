package controllers

import (
	"fmt"
)

var (
	DefaultRegistryHelper RegistryHelper = &registryHelper{
		dockerRegistryFQDN:            "function-controller-docker-registry.kyma-system.svc.cluster.local",
		dockerRegistryPort:            5000,
		dockerRegistryName:            "",
		dockerRegistryExternalAddress: "https://registry.kyma.local",
	}
)

type RegistryHelper interface {
	ImageName(name, namespace, tag string) string
	BuildImageName(name, namespace, tag string) string
	ServiceImageName(name, namespace, tag string) string
}

type registryHelper struct {
	dockerRegistryFQDN            string
	dockerRegistryPort            int
	dockerRegistryName            string
	dockerRegistryExternalAddress string
}

func (h *registryHelper) ImageName(name, namespace, tag string) string {
	if h.dockerRegistryName == "" {
		return fmt.Sprintf(
			"%s-%s:%s",
			name,
			namespace,
			tag,
		)
	}

	return fmt.Sprintf(
		"%s/%s-%s:%s",
		h.dockerRegistryName,
		name,
		namespace,
		tag,
	)
}

func (h *registryHelper) BuildImageName(name, namespace, tag string) string {
	imgName := h.ImageName(name, namespace, tag)
	return fmt.Sprintf(
		"%s:%d/%s",
		h.dockerRegistryFQDN,
		h.dockerRegistryPort,
		imgName,
	)
}

func (h *registryHelper) ServiceImageName(name, namespace, tag string) string {
	imgName := h.ImageName(name, namespace, tag)
	return fmt.Sprintf(
		"%s/%s",
		h.dockerRegistryExternalAddress,
		imgName,
	)
}
