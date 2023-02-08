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
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

func (s *HttpServer) listExperiments(c *gin.Context) {
	mode, ok := c.GetQuery("launch_mode")
	var chaosList []*core.Experiment
	var err error
	if ok {
		chaosList, err = s.exp.ListByLaunchMode(context.Background(), mode)
	} else {
		chaosList, err = s.exp.List(context.Background())
	}
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, chaosList)
}

func (s *HttpServer) listExperimentRuns(c *gin.Context) {
	uid := c.Param("uid")
	runsList, err := s.chaos.ExpRun.ListByExperimentUID(context.Background(), uid)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, runsList)
}
