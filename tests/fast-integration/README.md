# Fast integration tests

## Why?
We need fast, reliable integration tests to speed up Kyma development. The developer productivity is directly proportional to the number of iterations you can make per day/hour/minute. The iteration includes writing code/test, executing tests, refactoring, and executing test after refactoring, and the length of such iteration is a turnaround time. The goal is to decrease the minimal turnaround time from current 90 minutes to less than 10 minutes (ten times). Fast integration tests will solve the problem partially. Other initiatives that are executed in parallel are equally important: switching to k3s starting to reduce kubernetes provisioning time, and implement parallel installation of Kyma components.

## How to speed up integration tests?

The current integration testing flow looks like this:
- Build test image(s), push it, ~ 2min/image
- Deploy octopus, ~1 min
- Deploy test pod (test image), ~ 1min/image
- Sleep 20 seconds to wait for sidecar (in many tests)
- Deploy "test scene", ~1 min/image
- Execute the test, 5 sec/test
- Wait for test completion and collect results. ~1 min

The plan is to keep only 2 steps:
- Deploy "test scene", 1-2 minutes (one scene for all the tests)
- Execute the test, 5 sec/test

This way we can reduce testing phase from about 40 minutes to about 4 minutes). 

# Run tests locally

Requirements: node.js installed and KUBECONFIG pointing to the kubernetes cluster with Kyma installed. If you don't have Kyma yet, you can quickly run it locally using this project: https://github.com/kyma-incubator/local-kyma

Checkout Kyma project:
```
git clone git@github.com:kyma-project/kyma.git
```

Install dependencies:
```
cd kyma/tests/fast-integration
npm install
```

Execute the tests:
```
npm test
```


# FAQ

## Why don't you use Octopus?
Octopus is a great tool for running tests inside Kubernetes cluster in a declarative way. But it is not the right tool for fast integration testing. The goal is to execute the tests in 4 minutes. With Octopus you need 4 minutes or more before test even start (2 minutes to build test image and push it to docker registry, 1 minute to deploy Octopus, 1 minute to deploy the test pod). 

## Why tests are written in node.js not in Go?

Several reasons:
- no compilation time 
- consise syntax (handling json responses from api-server or our test fixtures)
- lighter dependencies (@kubernetes/client-node) 
- educational value for our customers who can read tests to learn how to use Kyma features (none of our customers write code in Go, they use JavaScript, Java or Python)
