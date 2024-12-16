package pki

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/c2micro/c2msrv/internal/constants"
	"github.com/c2micro/c2msrv/internal/ent"
	"github.com/c2micro/c2msrv/internal/ent/pki"
	"github.com/c2micro/c2msrv/internal/utils"

	"github.com/go-faster/errors"
)

// newCa генерация нового CA и сохранение в БД
func (cfg *CAConfigV1) newCa(ctx context.Context, tx *ent.Tx) (*ent.Pki, error) {
	var ca *ent.Pki
	// генерируем приват+паблик
	key, err := ecdsa.GenerateKey(utils.ChooseEC(constants.ECLength), rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "generate CA ECDSA private key")
	}
	pub := key.Public()

	// формируем значения для серта
	c := &x509.Certificate{
		SerialNumber: big.NewInt(cfg.Serial),
		Subject: pkix.Name{
			OrganizationalUnit: []string{cfg.Subject.OrganizationalUnit},
			Organization:       []string{cfg.Subject.Organization},
			Country:            []string{cfg.Subject.Country},
			CommonName:         cfg.Subject.CommonName,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(constants.CertValidTime, 0, 0),
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
	}
	// создаем self-signed серт для CA
	crt, err := x509.CreateCertificate(rand.Reader, c, c, pub, key)
	if err != nil {
		return nil, errors.Wrap(err, "generate CA certificate")
	}
	// формируем сертификат в PEM формате
	crtPem := new(bytes.Buffer)
	if err = pem.Encode(crtPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: crt,
	}); err != nil {
		return nil, errors.Wrap(err, "create CA PEM certificate")
	}

	// формируем ключ в PEM формате
	k, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "marshal CA ECDSA private key")
	}
	keyPem := new(bytes.Buffer)
	if err = pem.Encode(keyPem, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: k,
	}); err != nil {
		return nil, errors.Wrap(err, "create CA PEM key")
	}
	// записываем в БД
	ca, err = tx.Pki.
		Create().
		SetType(pki.TypeCa).
		SetCert(crtPem.Bytes()).
		SetKey(keyPem.Bytes()).
		Save(ctx)
	return ca, err
}

// newCert генерация сертификата для GRPC сервера листенеров
func (cfg *ListenerConfigV1) newCert(ctx context.Context, catls tls.Certificate, tx *ent.Tx, ip net.IP) (*ent.Pki, error) {
	var listener *ent.Pki
	// генерируем приват+паблик
	key, err := ecdsa.GenerateKey(utils.ChooseEC(constants.ECLength), rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "generate listener ECDSA private key")
	}
	pub := key.Public()

	// формируем значения для серта
	c := &x509.Certificate{
		SerialNumber: big.NewInt(cfg.Serial),
		Subject: pkix.Name{
			Organization: []string{cfg.Subject.Organization},
			Country:      []string{cfg.Subject.Country},
			Province:     []string{cfg.Subject.Province},
			Locality:     []string{cfg.Subject.Locality},
			CommonName:   cfg.Subject.CommonName,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(constants.CertValidTime, 0, 0),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		IPAddresses: []net.IP{ip},
	}
	ca, err := x509.ParseCertificate(catls.Certificate[0])
	if err != nil {
		return nil, errors.Wrap(err, "parse CA certificate")
	}
	// создаем self-signed серт для CA
	crt, err := x509.CreateCertificate(rand.Reader, c, ca, pub, catls.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "create listener certificate")
	}
	// формируем сертификат в PEM формате
	crtPem := new(bytes.Buffer)
	if err = pem.Encode(crtPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: crt,
	}); err != nil {
		return nil, err
	}

	// формируем ключ в PEM формате
	k, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "marshal listener ECDSA private key")
	}
	keyPem := new(bytes.Buffer)
	if err = pem.Encode(keyPem, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: k,
	}); err != nil {
		return nil, errors.Wrap(err, "create listener PEM key")
	}
	// записываем в БД
	listener, err = tx.Pki.
		Create().
		SetType(pki.TypeListener).
		SetCert(crtPem.Bytes()).
		SetKey(keyPem.Bytes()).
		Save(ctx)
	return listener, err
}

// newCert генерация сертификата для GRPC сервера операторов
func (cfg *OperatorConfigV1) newCert(ctx context.Context, catls tls.Certificate, tx *ent.Tx, ip net.IP) (*ent.Pki, error) {
	var listener *ent.Pki
	// генерируем приват+паблик
	key, err := ecdsa.GenerateKey(utils.ChooseEC(constants.ECLength), rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "generate listener ECDSA private key")
	}
	pub := key.Public()

	// формируем значения для серта
	c := &x509.Certificate{
		SerialNumber: big.NewInt(cfg.Serial),
		Subject: pkix.Name{
			Organization: []string{cfg.Subject.Organization},
			Country:      []string{cfg.Subject.Country},
			Province:     []string{cfg.Subject.Province},
			Locality:     []string{cfg.Subject.Locality},
			CommonName:   cfg.Subject.CommonName,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(constants.CertValidTime, 0, 0),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		IPAddresses: []net.IP{ip},
	}
	ca, err := x509.ParseCertificate(catls.Certificate[0])
	if err != nil {
		return nil, errors.Wrap(err, "parse CA certificate")
	}
	// создаем self-signed серт для CA
	crt, err := x509.CreateCertificate(rand.Reader, c, ca, pub, catls.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "create listener certificate")
	}
	// формируем сертификат в PEM формате
	crtPem := new(bytes.Buffer)
	if err = pem.Encode(crtPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: crt,
	}); err != nil {
		return nil, err
	}

	// формируем ключ в PEM формате
	k, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "marshal listener ECDSA private key")
	}
	keyPem := new(bytes.Buffer)
	if err = pem.Encode(keyPem, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: k,
	}); err != nil {
		return nil, errors.Wrap(err, "create listener PEM key")
	}
	// записываем в БД
	listener, err = tx.Pki.
		Create().
		SetType(pki.TypeOperator).
		SetCert(crtPem.Bytes()).
		SetKey(keyPem.Bytes()).
		Save(ctx)
	return listener, err
}

func (cfg *ManagementConfigV1) newCert(ctx context.Context, catls tls.Certificate, tx *ent.Tx, ip net.IP) (*ent.Pki, error) {
	var mgmt *ent.Pki
	// генерируем приват+паблик
	key, err := ecdsa.GenerateKey(utils.ChooseEC(constants.ECLength), rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "generate management ECDSA private key")
	}
	pub := key.Public()

	// формируем значения для серта
	c := &x509.Certificate{
		SerialNumber: big.NewInt(cfg.Serial),
		Subject: pkix.Name{
			Organization: []string{cfg.Subject.Organization},
			Country:      []string{cfg.Subject.Country},
			Province:     []string{cfg.Subject.Province},
			Locality:     []string{cfg.Subject.Locality},
			CommonName:   cfg.Subject.CommonName,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(constants.CertValidTime, 0, 0),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		IPAddresses: []net.IP{ip},
	}
	ca, err := x509.ParseCertificate(catls.Certificate[0])
	if err != nil {
		return nil, errors.Wrap(err, "parse CA certificate")
	}
	// создаем self-signed серт для CA
	crt, err := x509.CreateCertificate(rand.Reader, c, ca, pub, catls.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "create management certificate")
	}
	// формируем сертификат в PEM формате
	crtPem := new(bytes.Buffer)
	if err = pem.Encode(crtPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: crt,
	}); err != nil {
		return nil, err
	}

	// формируем ключ в PEM формате
	k, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "marshal management ECDSA private key")
	}
	keyPem := new(bytes.Buffer)
	if err = pem.Encode(keyPem, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: k,
	}); err != nil {
		return nil, errors.Wrap(err, "create management PEM key")
	}
	// записываем в БД
	mgmt, err = tx.Pki.
		Create().
		SetType(pki.TypeManagement).
		SetCert(crtPem.Bytes()).
		SetKey(keyPem.Bytes()).
		Save(ctx)
	return mgmt, err
}
