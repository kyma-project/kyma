package logging

import (
	log "github.com/sirupsen/logrus"
)

func GetLogger(tenant, group, id string) *log.Entry {
	return log.WithFields(log.Fields{
		"Group":  group,
		"Tenant": tenant,
		"ID":     id,
	})
}
