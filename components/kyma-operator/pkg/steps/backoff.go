package steps

import (
	"errors"
	"sync/atomic"
	"time"
)

//Implements a single backOff controller.
//It's role is to introduce configurable delays for repeating operations.
type backOffController struct {
	count      uint32
	intervals  []uint
	onStepFunc backOffStepFunc
	sleepFunc  func(seconds uint)
}

//Callback that runs on 'step(msg)' invocations
//count is equal to 0 on first call.
//max is the number equal to: len(intervals) - 1. If count > max all intervals have been used.
//delay is the delay configured for current iteration.
//msg is the argument passed to step() function
type backOffStepFunc func(count, max, delay int, msg ...string)

//Returns new backOffController
//intervals gives wait times for each 'step()' invocations.
//onStep is a Callback function
func newBackOff(intervals []uint, onStep backOffStepFunc) (*backOffController, error) {
	if len(intervals) < 1 {
		return nil, errors.New("Not enough intervals")
	}

	if onStep == nil {
		return nil, errors.New("onStep function is missing")
	}

	sleepFunc := func(seconds uint) {
		time.Sleep(time.Duration(seconds) * time.Second)
	}

	return &backOffController{
		count:      0,
		intervals:  intervals,
		onStepFunc: onStep,
		sleepFunc:  sleepFunc,
	}, nil
}

//Executes single backoff step.
//Upon each invocation this function first invokes configured callback function, then it blocks according to configured intervals.
func (bof *backOffController) step(msg ...string) {

	currCount := int(atomic.LoadUint32(&bof.count))

	cappedIdx := currCount
	if cappedIdx > len(bof.intervals)-1 {
		cappedIdx = len(bof.intervals) - 1
	}

	if bof.onStepFunc != nil {
		bof.onStepFunc(currCount, len(bof.intervals)-1, int(bof.intervals[cappedIdx]), msg...)
	}

	if bof.sleepFunc != nil {
		sleepSecs := bof.intervals[cappedIdx]
		bof.sleepFunc(sleepSecs)
	}

	atomic.AddUint32(&bof.count, 1)
}

func (bof *backOffController) limitReached() bool {
	currCount := int(atomic.LoadUint32(&bof.count))
	return currCount > (len(bof.intervals) - 1)
}

func (bof *backOffController) reset() {
	atomic.StoreUint32(&bof.count, 0)
}

func (bof *backOffController) currentCount() int {
	return int(atomic.LoadUint32(&bof.count))
}
