package certificates

import "fmt"

type CSRSubject struct {
	CommonName         string
	Country            string
	Organization       string
	OrganizationalUnit string
	Locality           string
	Province           string
}

func (s CSRSubject) ToString() string {
	return fmt.Sprintf("O=%s,OU=%s,L=%s,ST=%s,C=%s,CN=%s", s.Organization, s.OrganizationalUnit, s.Locality, s.Province, s.Country, s.CommonName)
}

type EncodedCertificateChain struct {
	CertificateChain  string
	ClientCertificate string
	CaCertificate     string
}
