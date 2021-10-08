// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package httpserver

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/server/utils"
)

const (
	MTLSServer = "mTLS"
	TLSServer  = "tls"
	HTTPServer = "http"
)

var (
	errMissingClientCert = utils.ErrAuth.New("Sorry, but you need to provide a client certificate to continue")
)

func (s *httpServer) serverMode() string {
	if len(s.conf.SSLCertFile) > 0 {
		if len(s.conf.SSLClientCAFile) > 0 {
			return MTLSServer
		}
		return TLSServer
	}
	return HTTPServer
}

func (s *httpServer) Run(handler func()) (err error) {
	addr := s.conf.Address()
	mode := s.serverMode()

	if mode == MTLSServer {
		log.Info("starting HTTPS server with Client Auth", zap.String("address", addr))

		caCert, ioErr := ioutil.ReadFile(s.conf.SSLClientCAFile)
		if ioErr != nil {
			err = ioErr
			return
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig := &tls.Config{
			ClientCAs:  caCertPool,
			ClientAuth: tls.RequestClientCert,
		}

		s.engine.Use(authenticateClientCert(tlsConfig))
		server := &http.Server{
			Addr:      addr,
			TLSConfig: tlsConfig,
			Handler:   s.engine,
		}

		handler()
		err = server.ListenAndServeTLS(s.conf.SSLCertFile, s.conf.SSLKeyFile)
	} else if mode == TLSServer {
		log.Info("starting HTTPS server", zap.String("address", addr))

		handler()
		err = s.engine.RunTLS(addr, s.conf.SSLCertFile, s.conf.SSLKeyFile)
	} else {
		log.Info("starting HTTP server", zap.String("address", addr))

		handler()
		err = s.engine.Run(addr)
	}
	return
}

func verifyCertificates(config *tls.Config, certs []*x509.Certificate) error {
	t := config.Time
	if t == nil {
		t = time.Now
	}
	opts := x509.VerifyOptions{
		Roots:         config.ClientCAs,
		CurrentTime:   t(),
		Intermediates: x509.NewCertPool(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	for _, cert := range certs[1:] {
		opts.Intermediates.AddCert(cert)
	}

	_, err := certs[0].Verify(opts)
	if err != nil {
		return err
	}
	return nil
}

func authenticateClientCert(config *tls.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientTLS := ctx.Request.TLS
		if len(clientTLS.PeerCertificates) > 0 {
			if err := verifyCertificates(config, clientTLS.PeerCertificates); err != nil {
				_ = ctx.AbortWithError(http.StatusForbidden, utils.ErrAuth.Wrap(err, "Unauthorized certificate credentials"))
			} else {
				ctx.Next()
			}
		} else {
			_ = ctx.AbortWithError(http.StatusUnauthorized, errMissingClientCert)
		}
	}
}
