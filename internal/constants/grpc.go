package constants

import (
	"crypto/tls"
	"time"
)

const (
	// отправка keepalive сообщения каждые 30 минут
	GrpcKeepaliveTime = time.Second * 30
	// время ожидания ответа на keepalive сообщение
	GrpcKeepaliveTimeout = time.Second * 5
	// время, после которого будет принудительно закрыто соединение
	GrpcMaxConnAgeGrace = time.Second * 10
	// минимальная версия TLS
	GrpcTlsMinVersion = tls.VersionTLS12
	// временная дельта для отправки heartbeat'ов оператору
	GrpcOperatorHealthCheckTimeout = time.Second * 10
)

var (
	// набор поддерживаемых шифров для TLS
	GrpcTlsCiphers = []uint16{
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	}
)
