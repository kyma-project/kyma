package certificates

type CSRSubject struct {
	CName              string
	Country            string
	Organization       string
	OrganizationalUnit string
	Locality           string
	Province           string
}
