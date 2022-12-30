package jetstreamv2

import (
	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
)

// Validate ensures that the NatsConfig is valid and therefore can be used safely.
// TODO: as soon as backend/nats is gone, make this method a function of backendnats.Config.
func Validate(natsConfig backendnats.Config) error {
	if natsConfig.JSStreamName == "" {
		return ErrEmptyStreamName
	}
	if len(natsConfig.JSStreamName) > jsMaxStreamNameLength {
		return ErrStreamNameTooLong
	}
	if _, err := toJetStreamStorageType(natsConfig.JSStreamStorageType); err != nil {
		return err
	}
	if _, err := toJetStreamRetentionPolicy(natsConfig.JSStreamRetentionPolicy); err != nil {
		return err
	}
	if _, err := toJetStreamDiscardPolicy(natsConfig.JSStreamDiscardPolicy); err != nil {
		return err
	}
	return nil
}
