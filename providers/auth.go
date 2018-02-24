package providers

type AuthBasic struct {
	User     string `hcl:"user"`
	Password string `hcl:"password"`
}

type TLS struct {
	TLSClientKey  string `hcl:"tls_client_key"`
	TLSClientCert string `hcl:"tls_client_cert"`
}

type AuthMTLS struct {
	TLSClientKey    string `hcl:"tls_client_key"`
	TLSClientCert   string `hcl:"tls_client_cert"`
	TLSClientCACert string `hcl:"tls_client_cacert"`
}
