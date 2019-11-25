package mapping

import v1 "k8s.io/api/core/v1"

func (c *Controller) ProcessItem(key string) error {
	return c.processItem(key)
}

func (c *Controller) DeleteAccessLabelFromNamespace(ns *v1.Namespace, name string) error {
	return c.ensureNsNotLabelled(ns, name)
}

func (c *Controller) GetAccessLabelFromApp(name string) (string, error) {
	return c.getAppAccLabel(name)
}
