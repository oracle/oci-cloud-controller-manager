package framework

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestTLSFixtureUsesSHA256Certificate(t *testing.T) {
	ca := parseTLSFixtureCertificate(t, SSLCAData)
	certificate := parseTLSFixtureCertificate(t, SSLCertificateData)

	if ca.SignatureAlgorithm != x509.SHA256WithRSA {
		t.Fatalf("CA uses %s; want %s", ca.SignatureAlgorithm, x509.SHA256WithRSA)
	}
	if certificate.SignatureAlgorithm != x509.SHA256WithRSA {
		t.Fatalf("certificate uses %s; want %s", certificate.SignatureAlgorithm, x509.SHA256WithRSA)
	}
	if !ca.IsCA {
		t.Fatal("TLS fixture CA is not a CA certificate")
	}

	keyBlock, _ := pem.Decode([]byte(SSLPrivateData))
	if keyBlock == nil {
		t.Fatal("failed to decode TLS fixture private key")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		t.Fatalf("parse TLS fixture private key: %v", err)
	}
	certificatePublicKey, err := x509.MarshalPKIXPublicKey(certificate.PublicKey)
	if err != nil {
		t.Fatalf("marshal TLS fixture certificate public key: %v", err)
	}
	privateKeyPublicKey, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal TLS fixture private-key public key: %v", err)
	}
	if !bytes.Equal(certificatePublicKey, privateKeyPublicKey) {
		t.Fatal("TLS fixture certificate and private key do not match")
	}

	roots := x509.NewCertPool()
	roots.AddCert(ca)
	if _, err := certificate.Verify(x509.VerifyOptions{
		Roots:     roots,
		DNSName:   "kubernetes.default.svc.cluster.local",
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}); err != nil {
		t.Fatalf("verify TLS fixture certificate chain: %v", err)
	}
}

func parseTLSFixtureCertificate(t *testing.T, certificateData string) *x509.Certificate {
	t.Helper()

	block, _ := pem.Decode([]byte(certificateData))
	if block == nil {
		t.Fatal("failed to decode TLS fixture certificate")
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse TLS fixture certificate: %v", err)
	}
	return certificate
}
