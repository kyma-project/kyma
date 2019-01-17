package certificates

import "crypto/x509/pkix"

type KeyCertPair struct {
	PrivateKey  []byte
	Certificate []byte
}

// TODO - probably can be removed
type Subject struct {
	CommonName         string
	Country            string
	Organization       string
	OrganizationalUnit string
	Locality           string
	Province           string
}

func (s Subject) asCertificateSubject() pkix.Name {
	return pkix.Name{
		CommonName:         s.CommonName,
		Country:            []string{s.Country},
		Organization:       []string{s.Organization},
		OrganizationalUnit: []string{s.OrganizationalUnit},
		Locality:           []string{s.Locality},
		Province:           []string{s.Province},
	}
}
