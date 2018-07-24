package tester

import (
	"fmt"
	"regexp"
	"sync"
	"time"
)

// Container represents information about container in pod
type Container struct {
	Ns            string
	PodName       string
	ContainerName string
}

// StateUpdate contains information about changes in container's State
type StateUpdate struct {
	Ready      bool
	RestartCnt int32
}

// State stores information about container's state
type State struct {
	Ready      bool
	RestartCnt int32
	ReadySince time.Time
}

// ContainerAndState aggregates Container and State
type ContainerAndState struct {
	Container Container
	State     State
}

// IgnoreContainersByRegexp provides ignore function which checks containers against patterns
func IgnoreContainersByRegexp(podPattern, nsPattern, containerPattern string) (IgnoringContainersFunc, error) {
	var podRegexp, nsRegexp, containerRegexp *regexp.Regexp
	var err error
	if podPattern != "" {
		podRegexp, err = regexp.Compile(podPattern)
		if err != nil {
			return nil, err
		}
	}
	if nsPattern != "" {
		nsRegexp, err = regexp.Compile(nsPattern)
		if err != nil {
			return nil, err
		}
	}

	if containerPattern != "" {
		containerRegexp, err = regexp.Compile(containerPattern)
		if err != nil {
			return nil, err
		}
	}

	return func(c Container) bool {
		match := true
		anyComparison := false
		if podRegexp != nil {
			match = match && podRegexp.MatchString(c.PodName)
			anyComparison = true
		}
		if nsRegexp != nil {
			match = match && nsRegexp.MatchString(c.Ns)
			anyComparison = true

		}

		if containerRegexp != nil {
			match = match && containerRegexp.MatchString(c.ContainerName)
			anyComparison = true
		}

		ignored := match && anyComparison
		if ignored {
			fmt.Printf("Ignoring container %+v\n", c)
		}
		return ignored
	}, nil

}

// IgnoringContainersFunc defines function for ignoring containers
type IgnoringContainersFunc func(c Container) bool

// NewClusterState creates ClusterState
func NewClusterState(ignoreFunc IgnoringContainersFunc) *ClusterState {
	if ignoreFunc == nil {
		ignoreFunc = func(c Container) bool {
			return false
		}
	}
	cs := &ClusterState{
		containers:         make(map[Container]State),
		timeProvider:       time.Now,
		isContainerIgnored: ignoreFunc,
	}

	return cs
}

// ClusterState contains information about all containers in cluster
type ClusterState struct {
	mtx                sync.Mutex
	containers         map[Container]State
	timeProvider       func() time.Time
	isContainerIgnored IgnoringContainersFunc
}

// ForgetPod removes all information about pod state
func (cs *ClusterState) ForgetPod(ns, podName string) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	for container := range cs.containers {
		if container.Ns == ns && container.PodName == podName {
			fmt.Printf("Forgetting about pod: [ns: %s, podName: %s]\n", ns, podName)
			delete(cs.containers, container)
		}
	}
}

// UpdateState update state of the container
func (cs *ClusterState) UpdateState(c Container, update StateUpdate) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	state, ex := cs.containers[c]

	if !ex {
		var since time.Time
		if update.Ready {
			since = cs.timeProvider()
		}
		state = State{
			Ready:      update.Ready,
			RestartCnt: update.RestartCnt,
			ReadySince: since,
		}
		cs.containers[c] = state
		return
	}

	var since time.Time
	if !update.Ready {
		since = time.Time{}
	} else {
		if state.Ready == false || update.RestartCnt != state.RestartCnt {
			since = cs.timeProvider()
		} else {
			since = state.ReadySince
		}

	}
	state.RestartCnt = update.RestartCnt
	state.Ready = update.Ready
	state.ReadySince = since

	cs.containers[c] = state

}

// GetUnstableContainers returns all Containers which are not ready or restarts in recent requiredStabilityPeriod
func (cs *ClusterState) GetUnstableContainers(requiredStabilityPeriod time.Duration) []ContainerAndState {
	cs.mtx.Lock()
	cs.mtx.Unlock()

	unstable := make([]ContainerAndState, 0)

	for c, s := range cs.containers {
		if cs.isContainerIgnored(c) {
			continue
		}
		if !s.Ready || cs.timeProvider().Sub(s.ReadySince) < requiredStabilityPeriod {
			unstable = append(unstable, ContainerAndState{
				Container: c,
				State:     s,
			})
		}
	}
	return unstable
}
