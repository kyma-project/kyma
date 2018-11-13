package v1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
)

type StatusCode int

func (c *StatusCode) String() string {
	return fmt.Sprintf("%T", c)
}

const (
	Empty StatusCode = iota
	InProgress
	Done
	Error
	HostnameOccupied
)

func (s StatusCode) IsEmpty() bool {
	return s == Empty
}

func (s StatusCode) IsInProgress() bool {
	return s == InProgress
}

func (s StatusCode) IsDone() bool {
	return s == Done
}

func (s StatusCode) IsError() bool {
	return s == Error
}

func (s StatusCode) IsHostnameOccupied() bool {
	return s == HostnameOccupied
}

type GatewayResourceStatus struct {
	Code      StatusCode      `json:"code"`
	Resource  GatewayResource `json:"resource,omitempty"`
	LastError string          `json:"lastError,omitempty"`
}

func (g *GatewayResourceStatus) String() string {
	return fmt.Sprintf("{Code: %v, Resource: %v, LastError: %v}", g.Code, g.Resource, g.LastError)
}

func (s *GatewayResourceStatus) IsEmpty() bool {
	return s.Code.IsEmpty()
}

func (s *GatewayResourceStatus) IsInProgress() bool {
	return s.Code.IsInProgress()
}

func (s *GatewayResourceStatus) IsDone() bool {
	return s.Code.IsDone()
}

func (s *GatewayResourceStatus) IsError() bool {
	return s.Code.IsError()
}

func (s *GatewayResourceStatus) IsHostnameOccupied() bool {
	return s.Code.IsHostnameOccupied()
}

type GatewayResource struct {
	Name    string    `json:"name,omitempty"`
	Version string    `json:"version,omitempty"`
	Uid     types.UID `json:"uid,omitempty"`
}

func (r *GatewayResource) String() string {
	return fmt.Sprintf("{Name: %s, Version: %s, Uid: %v}", r.Name, r.Version, r.Uid)
}
