package certificates

type Certificates struct {
	CRTChain  []byte
	ClientCRT []byte
	CaCRT     []byte
}
