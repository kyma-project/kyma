package mapping

import "k8s.io/api/core/v1"

func (c *Controller) ProcessItem(key string) error {
	return c.processItem(key)
}

func (c *Controller) DeleteAccessLabelFromNamespace(ns *v1.Namespace) error {
	return c.ensureNsNotLabelled(ns)
}

func (c *Controller) GetAccessLabelFromRE(name string) (string, error) {
	return c.getReAccLabel(name)
}
