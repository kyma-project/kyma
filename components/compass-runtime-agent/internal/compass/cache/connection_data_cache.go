package cache

import (
	"crypto/tls"

	"github.com/sirupsen/logrus"
)

type ConnectionData struct {
	Certificate  tls.Certificate
	DirectorURL  string
	ConnectorURL string
}

type ConnectionDataCache interface {
	AddSubscriber(s ConnectionDataSubscriber)
	UpdateConnectionData(cert tls.Certificate, directorURL, connectorURL string)
	UpdateURLs(directorURL, connectorURL string)
}

type connectionDataCache struct {
	connectionData ConnectionData
	subscribers    []ConnectionDataSubscriber
}

type ConnectionDataSubscriber func(data ConnectionData) error

func NewConnectionDataCache() *connectionDataCache {
	return &connectionDataCache{}
}

func (c *connectionDataCache) AddSubscriber(s ConnectionDataSubscriber) {
	c.subscribers = append(c.subscribers, s)
}

func (c *connectionDataCache) UpdateConnectionData(cert tls.Certificate, directorURL, connectorURL string) {
	c.connectionData.Certificate = cert
	c.connectionData.DirectorURL = directorURL
	c.connectionData.ConnectorURL = connectorURL

	c.notifySubscribers()
}

func (c *connectionDataCache) UpdateURLs(directorURL, connectorURL string) {
	c.connectionData.DirectorURL = directorURL
	c.connectionData.ConnectorURL = connectorURL

	c.notifySubscribers()
}

func (c *connectionDataCache) notifySubscribers() {
	for _, s := range c.subscribers {
		err := s(c.connectionData)
		if err != nil {
			logrus.Errorf("error notifying about connection data change: %s", err.Error())
		}
	}
}
