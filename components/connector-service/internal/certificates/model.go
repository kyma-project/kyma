package certificates

type CSRSubject struct {
	CommonName         string
	Country            string
	Organization       string
	OrganizationalUnit string
	Locality           string
	Province           string
}
