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

type ConvertingWebHook struct {
	scheme  *runtime.Scheme
	client  ctrlclient.Client
	decoder *conversion.Decoder
	log     logr.Logger
}

const (
	v1alpha1GitRepoNameAnnotation = "serverless.kyma-project.io/v1alpha1GitRepoName"
)

var _ http.Handler = &ConvertingWebHook{}

func NewConvertingWebHook(client ctrlclient.Client, scheme *runtime.Scheme) *ConvertingWebHook {
	//TODO: change signature of method scheme -> decoder
	decoder, _ := conversion.NewDecoder(scheme)
	log := controllerruntime.Log.WithName("Converting webhook")
	return &ConvertingWebHook{
		client:  client,
		scheme:  scheme,
		decoder: decoder,
		log:     log,
	}
}

func (w *ConvertingWebHook) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
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

func (w *ConvertingWebHook) handleConvertRequest(req *apix.ConversionRequest) (*apix.ConversionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("conversion request is nil")
	}
	var objects []runtime.RawExtension

	for _, obj := range req.Objects {
		src, gvk, err := w.decoder.Decode(obj.Raw)
		if err != nil {
			return nil, err
		}
		dst, err := w.allocateDstObject(req.DesiredAPIVersion, gvk.Kind)
		if err != nil {
			return nil, err
		}

		err = w.convertFunction(src, dst)
		if err != nil {
			return nil, err
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

func (w *ConvertingWebHook) allocateDstObject(apiVersion, kind string) (runtime.Object, error) {
	gvk := schema.FromAPIVersionAndKind(apiVersion, kind)

	obj, err := w.scheme.New(gvk)
	if err != nil {
		return obj, err
	}

	t, err := meta.TypeAccessor(obj)
	if err != nil {
		return obj, err
	}

	t.SetAPIVersion(apiVersion)
	t.SetKind(kind)

	return obj, nil
}

func (w *ConvertingWebHook) convertFunction(src, dst runtime.Object) error {
	// v1alpha1 -> v1alpha2
	if in, ok := src.(*serverlessv1alpha1.Function); ok {
		out, ok := dst.(*serverlessv1alpha2.Function)
		if !ok {
			dstGVK := dst.GetObjectKind().GroupVersionKind()
			return fmt.Errorf("unsupported convert destination version %s ", dstGVK)
		}
		out.ObjectMeta = in.ObjectMeta

		if in.Spec.Type == serverlessv1alpha1.SourceTypeGit {
			if out.Annotations != nil {
				out.Annotations[v1alpha1GitRepoNameAnnotation] = in.Spec.Source
			} else {
				out.Annotations = map[string]string{v1alpha1GitRepoNameAnnotation: in.Spec.Source}
			}
		}

		if err := w.convertSpecV1Alpha1ToV1Alpha2(&in.Spec, &out.Spec, in.Namespace); err != nil {
			return err
		}

		w.convertStatusV1Alpha1ToV1Alpha2(&in.Status, &out.Status)

		// v1alpha2 -> v1alpha1
	} else if in, ok := src.(*serverlessv1alpha2.Function); ok {
		out, ok := dst.(*serverlessv1alpha1.Function)
		if !ok {
			dstGVK := dst.GetObjectKind().GroupVersionKind()
			return fmt.Errorf("unsupported convert destination version %s ", dstGVK)
		}
		out.ObjectMeta = in.ObjectMeta
		repoName := ""
		if in.Annotations != nil {
			repoName = in.Annotations[v1alpha1GitRepoNameAnnotation]
		}
		if err := w.convertSpecV1Alpha2ToV1Alpha1(&in.Spec, &out.Spec, repoName, in.Name, in.Namespace); err != nil {
			return err
		}

		w.convertStatusV1Alpha2ToV1Alpha1(&in.Status, &out.Status)
	} else {
		dstGVK := dst.GetObjectKind().GroupVersionKind()
		return fmt.Errorf("unsupported convert source version %s ", dstGVK)
	}
	return nil
}

func (w *ConvertingWebHook) convertSpecV1Alpha1ToV1Alpha2(in *serverlessv1alpha1.FunctionSpec, out *serverlessv1alpha2.FunctionSpec, namespace string) error {
	out.Env = in.Env
	out.MaxReplicas = in.MaxReplicas
	out.MinReplicas = in.MaxReplicas
	out.Runtime = serverlessv1alpha2.Runtime(in.Runtime)
	out.RuntimeImageOverride = in.RuntimeImageOverride

	//TODO: out.ResourceConfiguration / profile
	out.ResourceConfiguration = serverlessv1alpha2.ResourceConfiguration{
		Build: serverlessv1alpha2.ResourceRequirements{
			Resources: in.BuildResources,
		},
		Function: serverlessv1alpha2.ResourceRequirements{
			Resources: in.Resources,
		},
	}
	out.Template = serverlessv1alpha2.Template{
		Labels: in.Labels,
	}
	if in.Type == serverlessv1alpha1.SourceTypeGit {
		repo := &serverlessv1alpha1.GitRepository{}
		err := w.client.Get(context.Background(),
			types.NamespacedName{
				Name:      in.Source,
				Namespace: namespace,
			}, repo)
		if err != nil {
			return err
		}

		out.Source = serverlessv1alpha2.Source{
			GitRepository: &serverlessv1alpha2.GitRepositorySource{
				URL: repo.Spec.URL,

				Repository: serverlessv1alpha2.Repository{
					BaseDir:   in.BaseDir,
					Reference: in.Reference,
				},
			},
		}

		if repo.Spec.Auth != nil {
			out.Source.GitRepository.Auth = &serverlessv1alpha2.RepositoryAuth{
				Type:       serverlessv1alpha2.RepositoryAuthType(repo.Spec.Auth.Type),
				SecretName: repo.Spec.Auth.SecretName,
			}
		}
	} else {
		out.Source = serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       in.Source,
				Dependencies: in.Deps,
			},
		}
	}
	return nil
}

func (w *ConvertingWebHook) convertStatusV1Alpha1ToV1Alpha2(in *serverlessv1alpha1.FunctionStatus, out *serverlessv1alpha2.FunctionStatus) {
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

func errored(err error) *apix.ConversionResponse {
	return &apix.ConversionResponse{
		Result: metav1.Status{
			Status:  metav1.StatusFailure,
			Message: err.Error(),
		},
	}
}

func (w *ConvertingWebHook) convertSpecV1Alpha2ToV1Alpha1(in *serverlessv1alpha2.FunctionSpec, out *serverlessv1alpha1.FunctionSpec, repoName, functionName, namespace string) error {
	out.Env = in.Env
	out.MaxReplicas = in.MaxReplicas
	out.MinReplicas = in.MaxReplicas
	out.Runtime = serverlessv1alpha1.Runtime(in.Runtime)
	out.RuntimeImageOverride = in.RuntimeImageOverride

	out.BuildResources = in.ResourceConfiguration.Build.Resources
	out.Resources = in.ResourceConfiguration.Function.Resources

	out.Labels = in.Template.Labels

	// TODO: clean check
	if in.Source.GitRepository.URL == "" {
		out.Source = in.Source.Inline.Source
		out.Deps = in.Source.Inline.Dependencies
	} else {
		// check repo name in the annotations,
		if repoName == "" {
			// create a random name git repo with the information. This is not supposed to happen, it's just a precaution.
			repo := &serverlessv1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: fmt.Sprintf("%s-", functionName),
					Namespace:    namespace,
				},
				Spec: serverlessv1alpha1.GitRepositorySpec{
					URL: in.Source.GitRepository.URL,
				},
			}
			if in.Source.GitRepository.Auth != nil {
				repo.Spec.Auth = &serverlessv1alpha1.RepositoryAuth{
					Type:       serverlessv1alpha1.RepositoryAuthType(in.Source.GitRepository.Auth.Type),
					SecretName: in.Source.GitRepository.Auth.SecretName,
				}
			}
			if err := w.client.Create(context.Background(), repo); err != nil {
				return errors.Wrap(err, "failed to create GitRepository")
			}
			repoName = repo.GetName()
		}
		out.Source = repoName
		out.Reference = in.Source.GitRepository.Reference
		out.BaseDir = in.Source.GitRepository.BaseDir
	}

	return nil
}

func (w *ConvertingWebHook) convertStatusV1Alpha2ToV1Alpha1(in *serverlessv1alpha2.FunctionStatus, out *serverlessv1alpha1.FunctionStatus) {
	out.Repository = serverlessv1alpha1.Repository(in.Repository)
	out.Commit = in.Commit
	out.Runtime = serverlessv1alpha1.RuntimeExtended(in.Runtime)
	out.RuntimeImageOverride = in.RuntimeImageOverride

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
