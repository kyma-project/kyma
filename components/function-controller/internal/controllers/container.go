package controllers

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

type Container struct {
	Manager ctrl.Manager
}
