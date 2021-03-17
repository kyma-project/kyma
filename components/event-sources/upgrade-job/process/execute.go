package process

import "github.com/pkg/errors"

func (p Process) Execute() error {
	for _, s := range p.Steps {
		err := s.Do()
		if err != nil {
			return errors.Wrapf(err, "failed in step: %s", s.ToString())
		}
	}
	return nil
}
