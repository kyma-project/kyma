package istio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/istio/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InitializeScenarioReconcilation(ctx *godog.ScenarioContext) {
	profile := os.Getenv(deployedKymaProfileVar)

	reconcilationCase := istioReconcilationCase{}
	reconcilationCase.command = helpers.Command{Cmd: "./kyma", Args: []string{"deploy", "-s", "main", "--component", "istio", "-v", "--ci", "-p", profile}, OutputChannel: make(chan string)}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		out, err := os.Create("kyma")
		if err != nil {
			return ctx, err
		}
		err = os.Chmod("./kyma", 0777)
		if err != nil {
			return ctx, err
		}
		defer out.Close()

		resp, err := http.Get(fmt.Sprintf("https://storage.googleapis.com/kyma-cli-unstable/kyma-%s", runtime.GOOS))
		if err != nil {
			return ctx, err
		}

		_, err = io.Copy(out, resp.Body)
		return ctx, err
	})

	ctx.Step(`^a reconcilation takes place$`, reconcilationCase.aReconcilationTakesPlace)
	ctx.Step(`^istioctl install takes place$`, reconcilationCase.istioctlInstallTakesPlace)
	ctx.Step(`^the httpbin deployment in "([^"]*)" namespace gets restarted until there is no sidecar$`, reconcilationCase.httpbinGetsRestartedUntilThereIsNoSidecar)
	ctx.Step(`^reconciler restarts the faulty deployment$`, reconcilationCase.reconcilerRestartsTheFaultyDeployment)
	InitializeScenarioIstioInstalled(ctx)
}

func (i *istioReconcilationCase) aReconcilationTakesPlace() error {
	waitChannel, err := i.command.Run()
	if err != nil {
		return err
	}
	i.doneChannel = waitChannel
	return nil
}

func (i *istioReconcilationCase) istioctlInstallTakesPlace() error {
	for {
		select {
		case line := <-i.command.OutputChannel:
			if strings.Contains(line, "Creating executable istioctl apply command") {
				return nil
			}
		case <-i.doneChannel:
			return errors.New("Istioctl install didn't take place")
		}
	}
}

func (i *istioReconcilationCase) httpbinGetsRestartedUntilThereIsNoSidecar(namespace string) error {
	noSidecar := make(chan bool)
	go func() {
		for {
			deployments, err := k8sClient.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
			if err != nil {
				break
			}
			for _, dep := range deployments.Items {
				if dep.Spec.Template.Annotations == nil {
					dep.Spec.Template.Annotations = map[string]string{}
				}
				dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().String()
				_, updateErr := k8sClient.AppsV1().Deployments(namespace).Update(context.Background(), &dep, metav1.UpdateOptions{})
				if updateErr != nil {
					break
				}
				time.Sleep(500 * time.Millisecond)

				pods, err := k8sClient.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
					LabelSelector: "app=httpbin",
				})
				if err != nil {
					break
				}
				for _, pod := range pods.Items {
					if !hasIstioProxy(pod.Spec.Containers) {
						noSidecar <- true
						return
					}
				}
			}
		}
	}()
	for {
		select {
		case <-i.doneChannel:
			return errors.New("The httpbin deployment could not be restarted to a state without sidecar")
		case <-noSidecar:
			time.Sleep(time.Millisecond * 200)
			return nil
		// Continue with reconcilation
		case <-i.command.OutputChannel:

		}
	}
}

func (i *istioReconcilationCase) reconcilerRestartsTheFaultyDeployment() error {
	for {
		select {
		case line := <-i.command.OutputChannel:
			if strings.Contains(line, "Proxy reset for 1 pods without sidecar successfully done") {
				return nil
			}
		case <-i.doneChannel:
			return errors.New("Reconcilation didn't restart the faulty pod")
		}
	}
}
