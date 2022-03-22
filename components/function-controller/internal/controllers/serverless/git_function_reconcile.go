package serverless

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path"
	"strings"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	fnRuntime "github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type gitFunctionReconciler struct {
	functionReconciler *FunctionReconciler
}

func newGitFunctionReconciler(fr *FunctionReconciler) *gitFunctionReconciler {
	return &gitFunctionReconciler{
		functionReconciler: fr,
	}
}

func (r *gitFunctionReconciler) reconcileGitFunction(ctx context.Context, instance *serverlessv1alpha1.Function, log logr.Logger) (ctrl.Result, error) {

	resources, err := r.functionReconciler.fetchFunctionResources(ctx, instance, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	dockerConfig, err := readDockerConfig(ctx, r.functionReconciler.client, r.functionReconciler.config, instance)
	if err != nil {
		log.Error(err, "Cannot read Docker registry configuration")
		return ctrl.Result{}, err
	}

	gitOptions, err := r.readGITOptions(ctx, instance)
	if err != nil {
		if updateErr := r.functionReconciler.updateStatusWithoutRepository(ctx, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionConfigurationReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonSourceUpdateFailed,
			Message:            fmt.Sprintf("Reading git options failed: %v", err),
		}); updateErr != nil {
			log.Error(err, "Reading git options failed")
			return ctrl.Result{}, errors.Wrap(updateErr, "while updating status")
		}
		return ctrl.Result{}, err
	}

	revision, err := r.syncRevision(instance, gitOptions)
	if err != nil {
		result, errMsg := NextRequeue(err)
		// TODO: This return masks the error from r.syncRevision() and doesn't pass it to the controller. This should be fixed in a follow up PR.
		return result, r.functionReconciler.updateStatusWithoutRepository(ctx, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionConfigurationReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonSourceUpdateFailed,
			Message:            errMsg,
		})
	}

	rtmCfg := fnRuntime.GetRuntimeConfig(instance.Spec.Runtime)
	var result ctrl.Result

	switch {
	case r.isOnSourceChange(instance, revision):
		return result, r.onSourceChange(ctx, instance, &serverlessv1alpha1.Repository{
			Reference: instance.Spec.Reference,
			BaseDir:   instance.Spec.Repository.BaseDir,
		}, revision)

	case r.isOnJobChange(instance, rtmCfg, resources.jobs.Items, resources.deployments.Items, gitOptions, dockerConfig):
		return r.onGitJobChange(ctx, log, instance, rtmCfg, resources.jobs.Items, gitOptions, dockerConfig)

	case r.functionReconciler.isOnDeploymentChange(instance, rtmCfg, resources.deployments.Items, dockerConfig):
		return r.functionReconciler.onDeploymentChange(ctx, log, instance, rtmCfg, resources.deployments.Items, dockerConfig)

	case r.functionReconciler.isOnServiceChange(instance, resources.services.Items):
		return result, r.functionReconciler.onServiceChange(ctx, log, instance, resources.services.Items)

	case r.functionReconciler.isOnHorizontalPodAutoscalerChange(instance, resources.hpas.Items, resources.deployments.Items):
		return result, r.functionReconciler.onHorizontalPodAutoscalerChange(ctx, log, instance, resources.hpas.Items, resources.deployments.Items[0].GetName())

	default:
		return r.functionReconciler.updateDeploymentStatus(ctx, log, instance, resources.deployments.Items, corev1.ConditionTrue)
	}
}

func (r *gitFunctionReconciler) syncRevision(instance *serverlessv1alpha1.Function, options git.Options) (string, error) {
	if instance.Spec.Type == serverlessv1alpha1.SourceTypeGit {
		return r.functionReconciler.gitOperator.LastCommit(options)
	}
	return "", nil
}

func NextRequeue(err error) (res ctrl.Result, errMsg string) {
	if git.IsNotRecoverableError(err) {
		return ctrl.Result{Requeue: false}, fmt.Sprintf("Stop reconciliation, reason: %s", err.Error())
	}

	errMsg = fmt.Sprintf("Sources update failed, reason: %v", err)
	if git.IsAuthErr(err) {
		errMsg = "Authorization to git server failed"
	}

	// use exponential delay
	return ctrl.Result{Requeue: true}, errMsg
}

func (r *gitFunctionReconciler) readGITOptions(ctx context.Context, instance *serverlessv1alpha1.Function) (git.Options, error) {
	if instance.Spec.Type != serverlessv1alpha1.SourceTypeGit {
		return git.Options{}, nil
	}

	var gitRepository serverlessv1alpha1.GitRepository
	if err := r.functionReconciler.client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: instance.Spec.Source}, &gitRepository); err != nil {
		return git.Options{}, errors.Wrap(err, "while getting git repository")
	}

	var auth *git.AuthOptions
	if gitRepository.Spec.Auth != nil {
		var secret corev1.Secret
		if err := r.functionReconciler.client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: gitRepository.Spec.Auth.SecretName}, &secret); err != nil {
			return git.Options{}, err
		}
		auth = &git.AuthOptions{
			Type:        git.RepositoryAuthType(gitRepository.Spec.Auth.Type),
			Credentials: readSecretData(secret.Data),
			SecretName:  gitRepository.Spec.Auth.SecretName,
		}
	}

	if instance.Spec.Reference == "" {
		return git.Options{}, fmt.Errorf("reference has to specified")
	}

	return git.Options{
		URL:       gitRepository.Spec.URL,
		Reference: instance.Spec.Reference,
		Auth:      auth,
	}, nil
}

func (r *gitFunctionReconciler) buildImageAddress(instance *serverlessv1alpha1.Function, registryAddress string) string {

	imageTag := r.calculateGitImageTag(instance)

	return fmt.Sprintf("%s/%s-%s:%s", registryAddress, instance.Namespace, instance.Name, imageTag)
}

func (r *gitFunctionReconciler) calculateGitImageTag(instance *serverlessv1alpha1.Function) string {
	data := strings.Join([]string{
		string(instance.GetUID()),
		instance.Status.Commit,
		instance.Status.Repository.BaseDir,
		string(instance.Status.Runtime),
	}, "-")
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func (r *gitFunctionReconciler) buildGitJob(instance *serverlessv1alpha1.Function, gitOptions git.Options, rtmConfig runtime.Config, dockerConfig DockerConfig) batchv1.Job {
	imageName := r.buildImageAddress(instance, dockerConfig.PushAddress)
	args := r.functionReconciler.config.Build.ExecutorArgs
	args = append(args, fmt.Sprintf("%s=%s", destinationArg, imageName), fmt.Sprintf("--context=dir://%s", workspaceMountPath))

	one := int32(1)
	zero := int32(0)
	rootUser := int64(0)
	optional := true

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-build-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
			Labels:       functionLabels(instance),
		},
		Spec: batchv1.JobSpec{
			Parallelism:           &one,
			Completions:           &one,
			ActiveDeadlineSeconds: nil,
			BackoffLimit:          &zero,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      functionLabels(instance),
					Annotations: istioSidecarInjectFalse,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "credentials",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: dockerConfig.ActiveRegistryConfigSecretName,
									Items: []corev1.KeyToPath{
										{
											Key:  ".dockerconfigjson",
											Path: ".docker/config.json",
										},
									},
								},
							},
						},
						{
							Name: "runtime",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: rtmConfig.DockerfileConfigMapName},
								},
							},
						},
						{
							Name:         "workspace",
							VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
						},
						{
							Name: "registry-config",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: r.functionReconciler.config.PackageRegistryConfigSecretName,
									Optional:   &optional,
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:            "repo-fetcher",
							Image:           r.functionReconciler.config.Build.RepoFetcherImage,
							Env:             buildRepoFetcherEnvVars(instance, gitOptions),
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workspace",
									MountPath: workspaceMountPath,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "executor",
							Image:           r.functionReconciler.config.Build.ExecutorImage,
							Args:            args,
							Resources:       instance.Spec.BuildResources,
							VolumeMounts:    r.getGitBuildJobVolumeMounts(instance, rtmConfig),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{Name: "DOCKER_CONFIG", Value: "/docker/.docker/"},
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: r.functionReconciler.config.BuildServiceAccountName,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &rootUser,
					},
				},
			},
		},
	}
}

func (r *gitFunctionReconciler) getGitBuildJobVolumeMounts(instance *serverlessv1alpha1.Function, rtmConfig runtime.Config) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{Name: "credentials", ReadOnly: true, MountPath: "/docker"},
		// Must be mounted with SubPath otherwise files are symlinks and it is not possible to use COPY in Dockerfile
		// If COPY is not used, then the cache will not work
		{Name: "workspace", MountPath: path.Join(workspaceMountPath, "src"), SubPath: strings.TrimPrefix(instance.Spec.BaseDir, "/")},
		{Name: "runtime", ReadOnly: true, MountPath: path.Join(workspaceMountPath, "Dockerfile"), SubPath: "Dockerfile"},
	}
	// add package registry config volume mount depending on the used runtime
	volumeMounts = append(volumeMounts, getPackageConfigVolumeMountsForRuntime(rtmConfig.Runtime)...)
	return volumeMounts
}

func (r *gitFunctionReconciler) isOnJobChange(instance *serverlessv1alpha1.Function, rtmCfg runtime.Config, jobs []batchv1.Job, deployments []appsv1.Deployment, gitOptions git.Options, dockerConfig DockerConfig) bool {
	image := r.buildImageAddress(instance, dockerConfig.PullAddress)
	buildStatus := getConditionStatus(instance.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)

	expectedJob := r.buildGitJob(instance, gitOptions, rtmCfg, dockerConfig)

	if len(deployments) == 1 &&
		deployments[0].Spec.Template.Spec.Containers[0].Image == image &&
		buildStatus != corev1.ConditionUnknown &&
		len(jobs) > 0 &&
		mapsEqual(expectedJob.GetLabels(), jobs[0].GetLabels()) {
		return buildStatus == corev1.ConditionFalse
	}

	return len(jobs) != 1 ||
		len(jobs[0].Spec.Template.Spec.Containers) != 1 ||
		// Compare image argument
		!equalJobs(jobs[0], expectedJob) ||
		!mapsEqual(expectedJob.GetLabels(), jobs[0].GetLabels()) ||
		buildStatus == corev1.ConditionUnknown ||
		buildStatus == corev1.ConditionFalse
}

func (r *gitFunctionReconciler) onGitJobChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, rtmCfg runtime.Config, jobs []batchv1.Job, gitOptions git.Options, dockerConfig DockerConfig) (ctrl.Result, error) {
	newJob := r.buildGitJob(instance, gitOptions, rtmCfg, dockerConfig)
	return r.functionReconciler.changeJob(ctx, log, instance, newJob, jobs)
}

func (r *gitFunctionReconciler) isOnSourceChange(instance *serverlessv1alpha1.Function, commit string) bool {
	return instance.Status.Commit == "" ||
		commit != instance.Status.Commit ||
		instance.Spec.Reference != instance.Status.Reference ||
		serverlessv1alpha1.RuntimeExtended(instance.Spec.Runtime) != instance.Status.Runtime ||
		instance.Spec.BaseDir != instance.Status.BaseDir ||
		getConditionStatus(instance.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady) == corev1.ConditionFalse
}

func (r *gitFunctionReconciler) onSourceChange(ctx context.Context, instance *serverlessv1alpha1.Function, repository *serverlessv1alpha1.Repository, commit string) error {
	return r.functionReconciler.updateStatus(ctx, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonSourceUpdated,
		Message:            fmt.Sprintf("Sources %s updated", instance.Name),
	}, repository, commit)
}
