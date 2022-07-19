package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/sirupsen/logrus"
	apix "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/conversion"
)

type ConvertingWebHook struct {
	scheme  *runtime.Scheme
	client  ctrlclient.Client
	decoder *conversion.Decoder
}

var _ http.Handler = &ConvertingWebHook{}

func NewConvertingWebHook(client ctrlclient.Client, scheme *runtime.Scheme) *ConvertingWebHook {
	//TODO: change signature of method scheme -> decoder
	decoder, _ := conversion.NewDecoder(scheme)
	return &ConvertingWebHook{
		client:  client,
		scheme:  scheme,
		decoder: decoder,
	}
}

// func (w *ConvertingWebHook) Handle(_ context.Context, req *apix.ConversionRequest) *apix.ConversionResponse {
// 	// if req.RequestKind.Kind == "Function" {

// 	return w.convertFunction(req)
// 	// }
// 	// return errored(fmt.Errorf("invalid kind: %v",))
// }

// func (w *ConvertingWebHook) InjectDecoder(decoder *conversion.Decoder) error {
// 	w.decoder = decoder
// 	return nil
// }

func (w *ConvertingWebHook) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	convertReview := &apix.ConversionReview{}
	buf, _ := ioutil.ReadAll(req.Body)

	err := json.NewDecoder(bytes.NewReader(buf)).Decode(convertReview)
	if err != nil {
		logrus.Error(err, "failed to read conversion request")
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	conversionResponse, err := w.handleConvertRequest(convertReview.Request)

	if err != nil {
		logrus.Error(err, "failed to convert", "request", convertReview.Request.UID)
		convertReview.Response = errored(err)
	} else {
		convertReview.Response = conversionResponse
	}
	convertReview.Response.UID = convertReview.Request.UID
	convertReview.Request = nil

	err = json.NewEncoder(resp).Encode(convertReview)
	if err != nil {
		logrus.Error(err, "failed to write response")
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
	in, ok := src.(*serverlessv1alpha1.Function)
	if !ok {
		srcGVK := src.GetObjectKind().GroupVersionKind()
		logrus.Warnf("unsupported convert source version %s ", srcGVK)
		return nil
	}

	out, ok := dst.(*serverlessv1alpha2.Function)
	if !ok {
		dstGVK := dst.GetObjectKind().GroupVersionKind()
		logrus.Warnf("unsupported convert destination version %s ", dstGVK)
		return nil
	}
	out.ObjectMeta = in.ObjectMeta
	//out.TypeMeta = in.TypeMeta

	if err := w.convertSpec(&in.Spec, &out.Spec, in.Namespace); err != nil {
		return err
	}

	w.convertStatus(&in.Status, &out.Status)

	//TODO: clean me???
	// fBytes, err := json.Marshal(out)
	// if err != nil {
	// 	return admission.Errored(http.StatusInternalServerError, err)
	// }
	// return admission.PatchResponseFromRaw(req.Object.Raw, fBytes)
	return nil
}

func (w *ConvertingWebHook) handleGitRepoDefaulting() admission.Response {
	return admission.Allowed("")
}

func (w *ConvertingWebHook) convertSpec(in *serverlessv1alpha1.FunctionSpec, out *serverlessv1alpha2.FunctionSpec, namespace string) error {
	out.Env = in.Env
	out.MaxReplicas = in.MaxReplicas
	out.MinReplicas = in.MaxReplicas
	out.Runtime = serverlessv1alpha2.Runtime(in.Runtime)
	out.RuntimeImageOverride = in.RuntimeImageOverride
	//out.ResourceConfiguration / profile
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
				Auth: &serverlessv1alpha2.RepositoryAuth{
					Type:       serverlessv1alpha2.RepositoryAuthType(repo.Spec.Auth.Type),
					SecretName: repo.Spec.Auth.SecretName,
				},
				Repository: serverlessv1alpha2.Repository{
					BaseDir:   in.BaseDir,
					Reference: in.Reference,
				},
			},
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

func (w *ConvertingWebHook) convertStatus(in *serverlessv1alpha1.FunctionStatus, out *serverlessv1alpha2.FunctionStatus) {
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
