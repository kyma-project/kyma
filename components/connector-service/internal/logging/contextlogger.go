package logging

import (
	log "github.com/sirupsen/logrus"
)

func GetClusterLogger(tenant, group string) *log.Entry {
	return log.WithFields(log.Fields{
		"Group":  group,
		"Tenant": tenant,
	})
}

func GetApplicationLogger(application, tenant, group string) *log.Entry {
	return GetClusterLogger(tenant, group).WithField("Application", application)
}
