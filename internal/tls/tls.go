package tls

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"

	"github.com/c2micro/c2m/internal/constants"
	"github.com/c2micro/c2m/internal/ent"
	"github.com/c2micro/c2m/internal/ent/pki"

	"github.com/go-faster/errors"
)

// NewTLSTransport создание новой конфигурации TLS транспорта
func NewTLSTransport(
	ctx context.Context,
	db *ent.Client,
	t pki.Type, // тип вытаскиваемого серта из БД
) (*tls.Config, string, error) {
	// получаем CA из БД
	ca, err := db.Pki.
		Query().
		Where(pki.TypeEQ(pki.TypeCa)).
		Only(ctx)
	if err != nil {
		return &tls.Config{}, "", errors.Wrap(err, "query CA key/cert from DB")
	}

	// получаем серт с заданным типом из БД
	v, err := db.Pki.
		Query().
		Where(pki.TypeEQ(t)).
		Only(ctx)
	if err != nil {
		return &tls.Config{}, "", errors.Wrapf(err, "query %s key/cert from DB", t.String())
	}

	// создаем x509 связки
	caX509, err := tls.X509KeyPair(ca.Cert, ca.Key)
	if err != nil {
		return &tls.Config{}, "", errors.Wrap(err, "create X509 keypair for CA")
	}
	vX509, err := tls.X509KeyPair(v.Cert, v.Key)
	if err != nil {
		return &tls.Config{}, "", errors.Wrapf(err, "create X509 keypair for %s", t.String())
	}

	// создание пула для CA сертов
	caPool := x509.NewCertPool()
	caCrt, err := x509.ParseCertificate(caX509.Certificate[0])
	if err != nil {
		return &tls.Config{}, "", errors.Wrap(err, "parse X509 certificate for CA")
	}
	// добавление CA серта в пул
	caPool.AddCert(caCrt)

	// вычисление fingerprint
	fp := sha1.Sum(vX509.Certificate[0])

	return &tls.Config{
		CipherSuites:           constants.GrpcTlsCiphers,
		MinVersion:             constants.GrpcTlsMinVersion,
		SessionTicketsDisabled: true,
		Certificates:           []tls.Certificate{vX509},
		ClientAuth:             tls.NoClientCert,
		RootCAs:                caPool,
	}, hex.EncodeToString(fp[:]), nil
}
