package v1alpha1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCompassConnection_ShouldRenewCertificate(t *testing.T) {

	testCases := []struct {
		renewNow                     bool
		certStatus                   *CertificateStatus
		minimalSyncTime              time.Duration
		certValidityRenewalThreshold float64
		shouldRenew                  bool
	}{
		{
			certStatus: &CertificateStatus{
				NotBefore: metav1.Now(),
				NotAfter:  metav1.NewTime(time.Now().Add(2000 * time.Hour)),
			},
			minimalSyncTime:              10 * time.Minute,
			certValidityRenewalThreshold: 0.3,
			shouldRenew:                  false,
		},
		{
			certStatus: &CertificateStatus{
				NotBefore: metav1.Now(),
				NotAfter:  metav1.Now(),
			},
			minimalSyncTime:              10 * time.Minute,
			certValidityRenewalThreshold: 0.3,
			shouldRenew:                  true,
		},
		{
			certStatus: &CertificateStatus{
				NotBefore: metav1.NewTime(time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)),
				NotAfter:  metav1.NewTime(time.Now().Add(3 * time.Hour)),
			},
			minimalSyncTime:              10 * time.Minute,
			shouldRenew:                  true,
			certValidityRenewalThreshold: 0.3,
		},
		{
			certStatus: &CertificateStatus{
				NotBefore: metav1.NewTime(time.Now()),
				NotAfter:  metav1.NewTime(time.Now().Add(30 * time.Minute)),
			},
			minimalSyncTime:              20 * time.Minute,
			shouldRenew:                  true,
			certValidityRenewalThreshold: 0.3,
		},
		{
			renewNow: true,
			certStatus: &CertificateStatus{
				NotBefore: metav1.Now(),
				NotAfter:  metav1.NewTime(time.Now().Add(2000 * time.Hour)),
			},
			minimalSyncTime:              10 * time.Minute,
			shouldRenew:                  true,
			certValidityRenewalThreshold: 0.3,
		},
	}

	for _, testCase := range testCases {
		connection := CompassConnection{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec:       CompassConnectionSpec{RefreshCredentialsNow: testCase.renewNow},
			Status: CompassConnectionStatus{
				ConnectionStatus: &ConnectionStatus{
					CertificateStatus: *testCase.certStatus,
				},
			},
		}

		willRenew := connection.ShouldRenewCertificate(testCase.certValidityRenewalThreshold, testCase.minimalSyncTime)

		assert.Equal(t, testCase.shouldRenew, willRenew)
	}

}
