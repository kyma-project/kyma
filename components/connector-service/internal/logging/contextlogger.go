package logging

import (
	log "github.com/sirupsen/logrus"
)

func GetClusterLogger(tenant, group, runtimeID string) *log.Entry {
	return log.WithFields(log.Fields{
		"Group":     group,
		"Tenant":    tenant,
		"RuntimeID": runtimeID,
	})
}

func GetApplicationLogger(application, tenant, group, runtimeID string) *log.Entry {
	return GetClusterLogger(tenant, group, runtimeID).WithField("Application", application)
}
