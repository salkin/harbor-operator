package harbor

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"time"
)

// helper function to create a cert template with a serial number and other required fields
func CertTemplate() (*x509.Certificate, error) {
	// generate a random serial number (a real cert authority would have some logic behind this)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.New("failed to generate serial number: " + err.Error())

	}

	tmpl := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{"Yhat, Inc."}},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour), // valid for an hour
		BasicConstraintsValid: true,
	}
	return &tmpl, nil

}

func CreateCert(template, parent *x509.Certificate, pub interface{}, parentPriv interface{}) (
	cert *x509.Certificate, certPEM []byte, err error) {

	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, pub, parentPriv)
	if err != nil {
		return
	}
	// parse the resulting certificate so we can use it again
	cert, err = x509.ParseCertificate(certDER)
	if err != nil {
		return
	}
	// PEM encode the certificate (this is a standard TLS encoding)
	b := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	certPEM = pem.EncodeToMemory(&b)
	return
}

func pemFromKey(key *rsa.PrivateKey) []byte {
	pemKey := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
	return pemKey
}

func createNewCertificate(rootKey *rsa.PrivateKey) ([]byte, *rsa.PrivateKey) {
	// generate a new key-pair
	servKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Error(err, "generating random key")
	}

	servCertTmpl, err := CertTemplate()
	if err != nil {
		log.Error(err, "creating cert template")
	}
	servCertTmpl.KeyUsage = x509.KeyUsageDigitalSignature
	servCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	servCertTmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}

	_, servCertPEM, err := CreateCert(servCertTmpl, servCertTmpl, &servKey.PublicKey, rootKey)
	if err != nil {
		log.Error(err, "error creating cert")
	}
	return servCertPEM, servKey
}

func createNewRoot(dnsNames []string) ([]byte, *rsa.PrivateKey) {
	// generate a new key-pair
	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Error(err, "generating random key")
	}

	rootCertTmpl, err := CertTemplate()
	if err != nil {
		log.Error(err, "creating cert template")
	}
	// describe what the certificate will be used for
	rootCertTmpl.IsCA = true
	rootCertTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
	rootCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	rootCertTmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	if len(dnsNames) > 0 {
		rootCertTmpl.DNSNames = dnsNames
	} else {
		rootCertTmpl.DNSNames = []string{"harbor"}
	}

	rootCert, rootCertPEM, err := CreateCert(rootCertTmpl, rootCertTmpl, &rootKey.PublicKey, rootKey)
	if err != nil {
		log.Error(err, "error creating cert")
	}
	// PEM encode the certificate (this is a standard TLS encoding)

	fmt.Printf("%s\n", rootCertPEM)
	fmt.Printf("%#x\n", rootCert.Signature) // more ugly binary

	return rootCertPEM, rootKey
}

func createHarbRoot() ([]byte, []byte) {
	// generate a new key-pair
	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Error(err, "generating random key")
	}

	rootCertTmpl, err := CertTemplate()
	if err != nil {
		log.Error(err, "creating cert template")
	}
	// describe what the certificate will be used for
	rootCertTmpl.IsCA = true
	rootCertTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
	rootCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	rootCertTmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	rootCertTmpl.DNSNames = []string{"harbor"}

	rootCert, rootCertPEM, err := CreateCert(rootCertTmpl, rootCertTmpl, &rootKey.PublicKey, rootKey)
	if err != nil {
		log.Error(err, "error creating cert")
	}
	// PEM encode the certificate (this is a standard TLS encoding)

	fmt.Printf("%s\n", rootCertPEM)
	fmt.Printf("%#x\n", rootCert.Signature) // more ugly binary

	rootPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(rootKey),
		},
	)
	return rootCertPEM, rootPem
}
