package validation

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConflictingFileName(t *testing.T) {
	l := telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Files: []telemetryv1alpha1.FileMount{{
				Name:    "labelMap.json",
				Content: "",
			},
			},
		},
	}
	pipeLineList := telemetryv1alpha1.LogPipelineList{}
	f := NewFilesValidator()
	err := f.Validate(&l, &pipeLineList)
	require.Error(t, err)
	require.Equal(t, "cannot use reserver filename 'labelmap.json'", err.Error())
}

func TestDuplicateFileName(t *testing.T) {
	l1 := telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Files: []telemetryv1alpha1.FileMount{{
				Name:    "f1.json",
				Content: "",
			},
			},
		},
	}

	pipeLineList := telemetryv1alpha1.LogPipelineList{}
	pipeLineList.Items = []telemetryv1alpha1.LogPipeline{l1}

	l2 := telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
		},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Files: []telemetryv1alpha1.FileMount{{
				Name:    "f1.json",
				Content: "",
			},
			},
		},
	}
	f := NewFilesValidator()
	err := f.Validate(&l2, &pipeLineList)
	require.Error(t, err)
	require.Equal(t, "filename 'f1.json' is already being used in the logPipeline 'foo'", err.Error())
}

func TestValidateUpdatePipeline(t *testing.T) {
	l1 := telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Files: []telemetryv1alpha1.FileMount{{
				Name:    "f1.json",
				Content: "",
			},
			},
		},
	}

	pipeLineList := telemetryv1alpha1.LogPipelineList{}
	pipeLineList.Items = []telemetryv1alpha1.LogPipeline{l1}

	l2 := telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Files: []telemetryv1alpha1.FileMount{{
				Name:    "f1.json",
				Content: "",
			},
			},
		},
	}
	f := NewFilesValidator()
	err := f.Validate(&l2, &pipeLineList)
	require.NoError(t, err)

}
