package process

// AddSteps defines the execution sequence of steps for upgrade-job
func (p *Process) AddSteps() {
	p.Steps = []Step{
		NewScaleDownEventingController(p),
		NewDeletePublisherDeployment(p),
		NewGetSubscriptions(p),
		NewFilterSubscriptions(p),
		NewDeleteBebSubscriptions(p),
	}
}