package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	apix "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/conversion"
)

type ConvertingWebhook struct {
	scheme  *runtime.Scheme
	client  ctrlclient.Client
	decoder *conversion.Decoder
	log     logr.Logger
}

const (
	v1alpha1GitRepoNameAnnotation = "serverless.kyma-project.io/v1alpha1GitRepoName"
)

var _ http.Handler = &ConvertingWebhook{}

func NewConvertingWebhook(client ctrlclient.Client, scheme *runtime.Scheme) *ConvertingWebhook {
	//TODO: change signature of method scheme -> decoder
	decoder, _ := conversion.NewDecoder(scheme)
	log := controllerruntime.Log.WithName("Converting webhook")
	return &ConvertingWebhook{
		client:  client,
		scheme:  scheme,
		decoder: decoder,
		log:     log,
	}
}

func (w *ConvertingWebhook) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	convertReview := &apix.ConversionReview{}
	err := json.NewDecoder(req.Body).Decode(convertReview)
	if err != nil {
		w.log.Error(err, "failed to read conversion request")
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	conversionResponse, err := w.handleConvertRequest(convertReview.Request)
	if err != nil {
		w.log.Error(err, "failed to convert", "request", convertReview.Request.UID)
		convertReview.Response = errored(err)
	} else {
		convertReview.Response = conversionResponse
	}
	convertReview.Response.UID = convertReview.Request.UID
	convertReview.Request = nil

	err = json.NewEncoder(resp).Encode(convertReview)
	if err != nil {
		w.log.Error(err, "failed to write response")
		return
	}
}

func (w *ConvertingWebhook) handleConvertRequest(req *apix.ConversionRequest) (*apix.ConversionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("conversion request is nil")
	}
	var objects []runtime.RawExtension

	for _, obj := range req.Objects {
		src, gvk, err := w.decoder.Decode(obj.Raw)
		if err != nil {
			return nil, errors.Wrap(err, "while decoding conversion request object")
		}
		dst, err := w.allocateDstObject(req.DesiredAPIVersion, gvk.Kind)
		if err != nil {
			return nil, errors.Wrap(err, "while allocating Dest object")
		}

		err = w.convertFunction(src, dst)
		if err != nil {
			return nil, errors.Wrap(err, "while applying function conversion")
		}
		objects = append(objects, runtime.RawExtension{Object: dst})
	}
	return &apix.ConversionResponse{
		UID:              req.UID,
		ConvertedObjects: objects,
		Result: metav1.Status{
			Status: metav1.StatusSuccess,
		},
	}, nil
}

func (w *ConvertingWebhook) allocateDstObject(apiVersion, kind string) (runtime.Object, error) {
	gvk := schema.FromAPIVersionAndKind(apiVersion, kind)

	obj, err := w.scheme.New(gvk)
	if err != nil {
		return obj, errors.Wrap(err, "while generating object")
	}

	t, err := meta.TypeAccessor(obj)
	if err != nil {
		return obj, errors.Wrap(err, "while accessing object type")
	}

	t.SetAPIVersion(apiVersion)
	t.SetKind(kind)

	return obj, nil
}

func (w *ConvertingWebhook) convertFunction(src, dst runtime.Object) error {
	switch src.(type) {
	// v1alpha1 -> v1alpha2
	case *serverlessv1alpha1.Function:
		return w.convertFunctionV1Alpha1ToV1Alpha2(src, dst)
	// v1alpha2 -> v1alpha1
	case *serverlessv1alpha2.Function:
		return w.convertFunctionV1Alpha2ToV1Alpha1(src, dst)
	default:
		dstGVK := dst.GetObjectKind().GroupVersionKind()
		return fmt.Errorf("unsupported convert source version %s ", dstGVK)
	}
}

// v1alpha1 -> v1alpha2
func (w *ConvertingWebhook) convertFunctionV1Alpha1ToV1Alpha2(src, dst runtime.Object) error {
	in := src.(*serverlessv1alpha1.Function)

	out, ok := dst.(*serverlessv1alpha2.Function)
	if !ok {
		dstGVK := dst.GetObjectKind().GroupVersionKind()
		return fmt.Errorf("unsupported convert destination version %s ", dstGVK)
	}

	out.ObjectMeta = in.ObjectMeta
	applyV1Alpha1ToV1Alpha2Annotations(in, out)

	if err := w.convertSpecV1Alpha1ToV1Alpha2(in, out); err != nil {
		return errors.Wrap(err, "while converting function spec from v1alpha1 to v1alpha2")
	}

	w.convertStatusV1Alpha1ToV1Alpha2(&in.Status, &out.Status)
	return nil
}

func applyV1Alpha1ToV1Alpha2Annotations(in *serverlessv1alpha1.Function, out *serverlessv1alpha2.Function) {
	if in.Spec.Type == serverlessv1alpha1.SourceTypeGit {
		if out.Annotations != nil {
			out.Annotations[v1alpha1GitRepoNameAnnotation] = in.Spec.Source
		} else {
			out.Annotations = map[string]string{v1alpha1GitRepoNameAnnotation: in.Spec.Source}
		}
	}
}

func (w *ConvertingWebhook) convertSpecV1Alpha1ToV1Alpha2(in *serverlessv1alpha1.Function, out *serverlessv1alpha2.Function) error {
	out.Spec.Env = in.Spec.Env
	if in.Spec.MinReplicas != nil || in.Spec.MaxReplicas != nil {
		out.Spec.ScaleConfig = &serverlessv1alpha2.ScaleConfig{
			MinReplicas: in.Spec.MinReplicas,
			MaxReplicas: in.Spec.MaxReplicas,
		}
	}
	out.Spec.Runtime = serverlessv1alpha2.Runtime(in.Spec.Runtime)
	out.Spec.RuntimeImageOverride = in.Spec.RuntimeImageOverride

	//TODO: out.Profile
	//TODO out.CustomRuntimeCOnfiguration

	out.Spec.ResourceConfiguration = serverlessv1alpha2.ResourceConfiguration{
		Build: serverlessv1alpha2.ResourceRequirements{
			Resources: in.Spec.BuildResources,
		},
		Function: serverlessv1alpha2.ResourceRequirements{
			Resources: in.Spec.Resources,
		},
	}
	out.Spec.Template = serverlessv1alpha2.Template{
		Labels: in.Spec.Labels,
	}
	if err := w.convertSourceV1Alpha1ToV1Alpha2(in, out); err != nil {
		return fmt.Errorf("failed to convert source from v1alpha1 to v1alpha2: %v", err)
	}
	return nil
}

func (w *ConvertingWebhook) convertSourceV1Alpha1ToV1Alpha2(in *serverlessv1alpha1.Function, out *serverlessv1alpha2.Function) error {
	if in.Spec.Type == serverlessv1alpha1.SourceTypeGit {
		return w.convertGitRepositoryV1Alpha1ToV1Alpha2(in, out)
	}
	out.Spec.Source = serverlessv1alpha2.Source{
		Inline: &serverlessv1alpha2.InlineSource{
			Source:       in.Spec.Source,
			Dependencies: in.Spec.Deps,
		},
	}
	return nil
}

func (w *ConvertingWebhook) convertGitRepositoryV1Alpha1ToV1Alpha2(in *serverlessv1alpha1.Function, out *serverlessv1alpha2.Function) error {
	repo := &serverlessv1alpha1.GitRepository{}
	err := w.client.Get(context.Background(),
		types.NamespacedName{
			Name:      in.Spec.Source,
			Namespace: in.Namespace,
		}, repo)
	if err != nil {
		return errors.Wrap(err, "while getting git repository")
	}
	out.Spec.Source = serverlessv1alpha2.Source{
		GitRepository: &serverlessv1alpha2.GitRepositorySource{
			URL: repo.Spec.URL,

			Repository: serverlessv1alpha2.Repository{
				BaseDir:   in.Spec.BaseDir,
				Reference: in.Spec.Reference,
			},
		},
	}

	if repo.Spec.Auth != nil {
		out.Spec.Source.GitRepository.Auth = &serverlessv1alpha2.RepositoryAuth{
			Type:       serverlessv1alpha2.RepositoryAuthType(repo.Spec.Auth.Type),
			SecretName: repo.Spec.Auth.SecretName,
		}
	}
	return nil
}

func (w *ConvertingWebhook) convertStatusV1Alpha1ToV1Alpha2(in *serverlessv1alpha1.FunctionStatus, out *serverlessv1alpha2.FunctionStatus) {
	out.Repository = serverlessv1alpha2.Repository(in.Repository)
	out.Commit = in.Commit
	out.Runtime = serverlessv1alpha2.RuntimeExtended(in.Runtime)
	out.RuntimeImageOverride = in.RuntimeImageOverride

	out.Conditions = []serverlessv1alpha2.Condition{}
	for _, c := range in.Conditions {
		out.Conditions = append(out.Conditions, serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionType(c.Type),
			Status:             c.Status,
			LastTransitionTime: c.LastTransitionTime,
			Reason:             serverlessv1alpha2.ConditionReason(c.Reason),
			Message:            c.Message,
		})
	}
}

// v1alpha2 -> v1alpha1
func (w *ConvertingWebhook) convertFunctionV1Alpha2ToV1Alpha1(src, dst runtime.Object) error {
	in := src.(*serverlessv1alpha2.Function)

	out, ok := dst.(*serverlessv1alpha1.Function)
	if !ok {
		dstGVK := dst.GetObjectKind().GroupVersionKind()
		return fmt.Errorf("unsupported convert destination version %s ", dstGVK)
	}

	out.ObjectMeta = in.ObjectMeta

	if err := w.convertSpecV1Alpha2ToV1Alpha1(in, out); err != nil {
		return errors.Wrap(err, "while converting function spec from v1alpha2 to v1alpha1")
	}

	w.convertStatusV1Alpha2ToV1Alpha1(&in.Status, out.Spec.Source, &out.Status)
	return nil
}

func (w *ConvertingWebhook) convertSpecV1Alpha2ToV1Alpha1(in *serverlessv1alpha2.Function, out *serverlessv1alpha1.Function) error {
	out.Spec.Env = in.Spec.Env
	out.Spec.Runtime = serverlessv1alpha1.Runtime(in.Spec.Runtime)
	out.Spec.RuntimeImageOverride = in.Spec.RuntimeImageOverride
	if in.Spec.ScaleConfig != nil {
		out.Spec.MinReplicas = in.Spec.ScaleConfig.MinReplicas
		out.Spec.MaxReplicas = in.Spec.ScaleConfig.MaxReplicas
	}

	out.Spec.BuildResources = in.Spec.ResourceConfiguration.Build.Resources
	out.Spec.Resources = in.Spec.ResourceConfiguration.Function.Resources

	out.Spec.Labels = in.Spec.Template.Labels

	return w.convertSourceV1Alpha2ToV1Alpha1(in, out)
}

func (w *ConvertingWebhook) convertSourceV1Alpha2ToV1Alpha1(in *serverlessv1alpha2.Function, out *serverlessv1alpha1.Function) error {
	if in.Spec.Source.Inline != nil {
		out.Spec.Source = in.Spec.Source.Inline.Source
		out.Spec.Deps = in.Spec.Source.Inline.Dependencies
		return nil
	}
	out.Spec.Type = serverlessv1alpha1.SourceTypeGit

	// check repo name in the annotations,
	// if not exists it means that the function was created as v1alpha2
	repoName := ""
	if in.Annotations != nil {
		repoName = in.Annotations[v1alpha1GitRepoNameAnnotation]
	}

	out.Spec.Source = repoName
	out.Spec.Reference = in.Spec.Source.GitRepository.Reference
	out.Spec.BaseDir = in.Spec.Source.GitRepository.BaseDir
	return nil
}

func (w *ConvertingWebhook) convertStatusV1Alpha2ToV1Alpha1(in *serverlessv1alpha2.FunctionStatus, outSource string, out *serverlessv1alpha1.FunctionStatus) {
	out.Repository = serverlessv1alpha1.Repository(in.Repository)
	out.Commit = in.Commit
	out.Runtime = serverlessv1alpha1.RuntimeExtended(in.Runtime)
	out.RuntimeImageOverride = in.RuntimeImageOverride
	out.Source = outSource

	out.Conditions = []serverlessv1alpha1.Condition{}

	for _, c := range in.Conditions {
		out.Conditions = append(out.Conditions, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionType(c.Type),
			Status:             c.Status,
			LastTransitionTime: c.LastTransitionTime,
			Reason:             serverlessv1alpha1.ConditionReason(c.Reason),
			Message:            c.Message,
		})
	}
}

func errored(err error) *apix.ConversionResponse {
	return &apix.ConversionResponse{
		Result: metav1.Status{
			Status:  metav1.StatusFailure,
			Message: err.Error(),
		},
	}
}
