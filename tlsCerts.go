package egressproxy

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	mathrand "math/rand"
	"os"
	"time"
)

var (
	caKey  string
	caCert string
	caPair tls.Certificate
)

func init() {
	var ok bool
	var err error
	var val string

	envVars := []string{
		"CA_CERT",
		"CA_KEY",
		"CA_COUNTRY",
		"CA_PROVINCE",
		"CA_LOCALITY",
		"CA_O",
		"CA_OU",
	}

	for _, v := range envVars {
		if val, ok = os.LookupEnv(v); !ok {
			log.Fatalf("missing environment variable: %v\n", v)
		}
		if val == "" {
			log.Fatalf("env var is set but empty: %v\n", v)
		}
	}

	caCert = os.Getenv("CA_CERT")
	caKey = os.Getenv("CA_KEY")
	caPair, err = tls.X509KeyPair([]byte(caCert), []byte(caKey))
	if err != nil {
		log.Fatalf("error loading CA key pair: %v", err)
	}
}

// makeCert creates a new TLS certificate, and signs it
func makeCert(hostname string) ([]byte, []byte) {
	caCert, _ := x509.ParseCertificate(caPair.Certificate[0])

	certPkix := pkix.Name{
		CommonName:         hostname,
		Country:            []string{os.Getenv("CA_COUNTRY")},
		Province:           []string{os.Getenv("CA_PROVINCE")},
		Locality:           []string{os.Getenv("CA_LOCALITY")},
		Organization:       []string{os.Getenv("CA_O")},
		OrganizationalUnit: []string{os.Getenv("CA_OU")},
	}

	certSerial := &big.Int{}
	randSeed := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))

	// populate cert template
	template := x509.Certificate{
		DNSNames:           []string{hostname},
		NotBefore:          time.Now(),
		NotAfter:           time.Now().Add(1 * time.Hour),
		Subject:            certPkix,
		SignatureAlgorithm: x509.PureEd25519, // the CA keypair must use ed25519 private key for this to work
		SerialNumber:       certSerial.SetInt64(randSeed.Int63()),
	}

	// create ed25519 key pair - don't check for errors e.g. missing /dev/urandom
	certPubKey, certPrivKey, _ := ed25519.GenerateKey(rand.Reader)

	// encode the private key
	certPrivKeyBuf := new(bytes.Buffer)
	certPrivKeyBytes, err := x509.MarshalPKCS8PrivateKey(certPrivKey)
	if err != nil {
		panic(err)
	}

	// don't check for errors since this writes to the newly allocated bytes.Buffer
	_ = pem.Encode(certPrivKeyBuf, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: certPrivKeyBytes,
	})

	// create the TLS certificate and sign using the CA cert
	certBytes, err := x509.CreateCertificate(
		rand.Reader,       // crypto rand.reader
		&template,         // populated certificate fields
		caCert,            // the ca public certificate
		certPubKey,        // sign the public key with the ca private key
		caPair.PrivateKey, // the ca private key
	)
	if err != nil {
		panic(err)
	}

	// PEM encode the TLS certificate
	certPEMBuf := new(bytes.Buffer)
	_ = pem.Encode(certPEMBuf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	return certPEMBuf.Bytes(), certPrivKeyBuf.Bytes()
}
