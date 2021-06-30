#!/usr/bin/env bash

# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# See the License for the specific language governing permissions and
# limitations under the License.

# Generate script because certs expire in 1 year (365 days)

mkdir -p client server

# generate server certificate
openssl req \
	-x509 \
	-newkey rsa:4096 \
	-keyout server/server_key.pem \
	-out server/server_cert.pem \
	-nodes \
	-days 365 \
	-subj "/CN=localhost/O=Client\ Certificate\ Demo"

# generate server-signed (valid) certifcate
openssl req \
	-newkey rsa:4096 \
	-keyout client/valid_key.pem \
	-out client/valid_csr.pem \
	-nodes \
	-days 365 \
	-subj "/CN=Valid"

# sign with server_cert.pem
openssl x509 \
	-req \
	-in client/valid_csr.pem \
	-CA server/server_cert.pem \
	-CAkey server/server_key.pem \
	-out client/valid_cert.pem \
	-set_serial 01 \
	-days 365

# generate self-signed (invalid) certifcate
openssl req \
	-newkey rsa:4096 \
	-keyout client/invalid_key.pem \
	-out client/invalid_csr.pem \
	-nodes \
	-days 365 \
	-subj "/CN=Invalid"

# sign with invalid_csr.pem
openssl x509 \
	-req \
	-in client/invalid_csr.pem \
	-signkey client/invalid_key.pem \
	-out client/invalid_cert.pem \
	-days 365
