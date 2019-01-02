package tokens

type TokenData struct {
	Group  string
	Tenant string
	Token  string
}

type Group struct {
	ID          string
	DisplayName string
}
