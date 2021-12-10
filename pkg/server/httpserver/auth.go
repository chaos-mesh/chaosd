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

func (s *httpServer) startHttpsServer() (err error) {
	mode := s.serverMode()
	if mode == HTTPServer {
		return nil
	}

	httpsServerAddr := s.conf.HttpsServerAddress()
	e := gin.Default()
	e.Use(utils.MWHandleErrors())
	s.handler(e)

	if mode == MTLSServer {
		log.Info("starting HTTPS server with Client Auth", zap.String("address", httpsServerAddr))

		caCert, ioErr := ioutil.ReadFile(s.conf.SSLClientCAFile)
		if ioErr != nil {
			err = ioErr
			return
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig := &tls.Config{
			ClientCAs:  caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
			ServerName: s.conf.ServerName,
		}

		server := &http.Server{
			Addr:      httpsServerAddr,
			TLSConfig: tlsConfig,
			Handler:   e,
		}

		err = server.ListenAndServeTLS(s.conf.SSLCertFile, s.conf.SSLKeyFile)
	} else if mode == TLSServer {
		log.Info("starting HTTPS server", zap.String("address", httpsServerAddr))
		err = e.RunTLS(httpsServerAddr, s.conf.SSLCertFile, s.conf.SSLKeyFile)
	}
	return
}
