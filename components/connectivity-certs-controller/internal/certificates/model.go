package certificates

type Certificates struct {
	ClientKey []byte
	CRTChain  []byte
	ClientCRT []byte
	CaCRT     []byte
}
