package process

func (p *Process) AddSteps() {
	p.Steps = []Step{
		NewSaveCurrentState(p),
		NewDeleteEventServiceDeployments(p),
		NewDeleteEventSources(p),
		NewUnLabelNamespace(p),
		NewDeleteBrokers(p),
		NewConvertTriggersToKymaSubscriptions(p),
		NewDeleteTriggers(p),
		NewDeleteAppSubscriptions(p),
		NewPatchConnectivityValidators(p),
		NewCleanUp(p),
	}
}
