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
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaos-mesh/chaosd/pkg/version"
)

type healthInfo struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (s *httpServer) healthcheck(c *gin.Context) {
	c.JSON(http.StatusOK, healthInfo{Status: 0})
}

func (s *httpServer) version(c *gin.Context) {
	c.JSON(http.StatusOK, version.Get())
}
