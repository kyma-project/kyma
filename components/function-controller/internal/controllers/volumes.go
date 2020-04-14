package controllers

import (
	"github.com/gogo/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
)

type configmapVolumeSpec struct {
	name string
	path string
	cmap string
}

const (
	defaultFileMode int32 = 420
)

func (s *configmapVolumeSpec) volume() *corev1.Volume {
	return &corev1.Volume{
		Name: s.name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				DefaultMode: proto.Int32(defaultFileMode),
				LocalObjectReference: corev1.LocalObjectReference{
					Name: s.cmap,
				},
			},
		},
	}
}

func (s *configmapVolumeSpec) volumeMount() *corev1.VolumeMount {
	return &corev1.VolumeMount{
		Name:      s.name,
		MountPath: s.path,
	}
}
