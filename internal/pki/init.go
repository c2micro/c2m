package pki

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/c2micro/c2msrv/internal/ent"
	"github.com/c2micro/c2msrv/internal/ent/pki"

	"github.com/go-faster/errors"
)

// Init инициализация PKI (CA + сертификаты для GRPC серверов)
func (c *ConfigV1) Init(ctx context.Context, db *ent.Client, lIP, oIP, mIP net.IP) error {
	var ca *ent.Pki
	var err error

	// создаем транзакцию
	tx, err := db.Tx(ctx)
	if err != nil {
		return errors.Wrap(err, "unable create tx")
	}

	// ca
	ca, err = tx.Pki.
		Query().
		Where(pki.TypeEQ(pki.TypeCa)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// генерация нового CA
			ca, err = c.CA.newCa(ctx, tx)
			if err != nil {
				return errors.Wrap(err, "unable create CA")
			}
			// если мы только создали CA - удаляем "возможно существующие" серты для listener/operator
			_, err = tx.Pki.Delete().Where(pki.TypeIn(pki.TypeListener, pki.TypeOperator)).Exec(ctx)
			if err != nil {
				return errors.Wrap(err, "unable remove listener and operator key/cert from DB")
			}
		} else {
			return errors.Wrap(err, "query CA's pki")
		}
	}

	caTLS, err := tls.X509KeyPair(ca.Cert, ca.Key)
	if err != nil {
		return errors.Wrap(err, "create CA x509 keypair")
	}

	// listener
	_, err = tx.Pki.
		Query().
		Where(pki.TypeEQ(pki.TypeListener)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// генерация нового сертификата для GRPC сервера листенеров
			_, err = c.Listener.newCert(ctx, caTLS, tx, lIP)
			if err != nil {
				return errors.Wrap(err, "create listener's key-cert")
			}
		} else {
			return errors.Wrap(err, "query listener's pki")
		}
	}

	// operator
	_, err = tx.Pki.
		Query().
		Where(pki.TypeEQ(pki.TypeOperator)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// генерация нового сертификата для GRPC сервера операторов
			_, err = c.Operator.newCert(ctx, caTLS, tx, oIP)
			if err != nil {
				return errors.Wrap(err, "create operator's key-cert")
			}
		} else {
			return errors.Wrap(err, "query operator's pki")
		}
	}

	// management
	_, err = tx.Pki.
		Query().
		Where(pki.TypeEQ(pki.TypeManagement)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// генерация нового сертификата для management GRPC сервера
			_, err = c.Management.newCert(ctx, caTLS, tx, mIP)
			if err != nil {
				return errors.Wrap(err, "create management key/cert")
			}
		} else {
			return errors.Wrap(err, "query management's key/cert")
		}
	}

	// коммит транзакции
	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "unable commit tx")
	}
	return nil
}
