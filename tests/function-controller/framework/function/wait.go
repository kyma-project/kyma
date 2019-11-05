/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package function

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/kubernetes/test/e2e/framework"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

const pollTimeout = time.Minute * 8

// WaitForStatusTransitions creates a watch on the given Function and evaluates
// its status transitions in the given order.
func WaitForStatusTransitions(fnUnstr *unstructured.Unstructured, cli dynamic.ResourceInterface,
	expectConds []serverlessv1alpha1.FunctionCondition) {

	fnName := fnUnstr.GetName()

	for _, c := range expectConds {
		ev := watchUntil(cli, fnUnstr.GetResourceVersion(),
			hasCondition(fnName, c),
		)
		fnUnstr = ev.Object.(*unstructured.Unstructured)
	}
}

// watchUntil evaluates conditions against a watched resource until one of them
// returns true, or after a timeout.
func watchUntil(cli dynamic.ResourceInterface, initialResourceVersion string, conds ...conditionFunc) *watch.Event {
	var condFuncs []watchtools.ConditionFunc
	for _, c := range conds {
		condFuncs = append(condFuncs, c.ConditionFunc)
	}

	ctx, cancel := context.WithTimeout(context.Background(), pollTimeout)
	defer cancel()

	ev, err := watchtools.Until(ctx, initialResourceVersion, cli, condFuncs...)
	if err != nil {
		framework.Failf("Error validating watch conditions %q: %v", conds, err)
	}

	return ev
}

// conditionFunc wraps a watch condition together with a description to make
// error reporting clearer.
type conditionFunc struct {
	watchtools.ConditionFunc
	description string
}

// String implements the Stringer interface.
func (f conditionFunc) String() string {
	return f.description
}

// hasCondition returns a watch condition that compares the status of a given
// Function with an expected value.
func hasCondition(fnName string, c serverlessv1alpha1.FunctionCondition) conditionFunc {
	checkFn := func(e watch.Event) (bool, error) {
		fnUnstr := e.Object.(*unstructured.Unstructured)
		if fnUnstr.GetName() != fnName {
			return false, nil
		}

		fn := FromUnstructured(fnUnstr)
		currentCond := fn.Status.Condition

		// fail fast if the Function errors unexpectedly
		if c != serverlessv1alpha1.FunctionConditionError &&
			currentCond == serverlessv1alpha1.FunctionConditionError {

			return false, fmt.Errorf("function in error condition")
		}

		return currentCond == c, nil
	}

	return conditionFunc{
		description:   fmt.Sprintf("Condition equals %s", c),
		ConditionFunc: checkFn,
	}
}
