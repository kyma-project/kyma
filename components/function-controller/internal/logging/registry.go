package logging

import (
	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Registry is an structure that allows to modify logLevel and logFormat
// of the main logger and sub-loggers on the fly
type Registry struct {
	logger           *logger.Logger
	desugaredLoggers []*zap.Logger
	namedLoggers     map[string]*zap.SugaredLogger
}

// ConfigureRegisteredLogger - create new registry
func ConfigureRegisteredLogger(logLevel, logFormat string) (*Registry, error) {
	r := &Registry{
		namedLoggers: map[string]*zap.SugaredLogger{},
	}
	return r.Reconfigure(logLevel, logFormat)
}

// CreateNamed - create and register zap.SugaredLogger. Sub-loggers of the created one would be not registered
func (r *Registry) CreateNamed(name string) *zap.SugaredLogger {
	l := r.createNamed(name)
	r.namedLoggers[name] = l
	return l
}

// CreateDesugared - create and register zap.Logger. Sub-loggers of the created one would be not registered
func (r *Registry) CreateDesugared() *zap.Logger {
	l := r.createDesugared()
	r.desugaredLoggers = append(r.desugaredLoggers, l)
	return l
}

// CreateUnregistered - create zap.Logger without registration
func (r *Registry) CreateUnregistered() *zap.SugaredLogger {
	return r.logger.WithContext()
}

// Reconfigure - apply new configuration for the logger and all registered sub-loggers
func (r *Registry) Reconfigure(logLevel, logFormat string) (*Registry, error) {
	log, err := ConfigureLogger(logLevel, logFormat)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure logger")
	}

	r.logger = log
	for key := range r.namedLoggers {
		*r.namedLoggers[key] = *r.createNamed(key)
	}

	for i := range r.desugaredLoggers {
		*r.desugaredLoggers[i] = *r.createDesugared()
	}

	return r, nil
}

func (r *Registry) createNamed(name string) *zap.SugaredLogger {
	l := r.logger.WithContext().Named(name).With()

	return l
}

func (r *Registry) createDesugared() *zap.Logger {
	l := r.logger.WithContext().Desugar()

	return l
}
