// Copyright 2020 Chaos Mesh Authors.
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
	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/config"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/server/utils"
	"github.com/chaos-mesh/chaosd/pkg/swaggerserver"
)

type httpServer struct {
	conf   *config.Config
	chaos  *chaosd.Server
	exp    core.ExperimentStore
	engine *gin.Engine
}

func NewServer(
	conf *config.Config,
	chaos *chaosd.Server,
	exp core.ExperimentStore,
) *httpServer {
	e := gin.Default()
	e.Use(utils.MWHandleErrors())

	return &httpServer{
		conf:   conf,
		chaos:  chaos,
		exp:    exp,
		engine: e,
	}
}

func Register(s *httpServer) {
	if s.conf.Platform != config.LocalPlatform {
		return
	}

	handler(s)

	go func() {
		addr := s.conf.Address()
		log.Debug("starting HTTP server", zap.String("address", addr))

		if err := s.engine.Run(addr); err != nil {
			log.Fatal("failed to start HTTP server", zap.Error(err))
		}
	}()
}

func handler(s *httpServer) {
	api := s.engine.Group("/api")
	{
		api.GET("/swagger/*any", swaggerserver.Handler())
	}

	attack := api.Group("/attack")
	{
		attack.POST("/process", s.createProcessAttack)
		attack.POST("/stress", s.createStressAttack)

		attack.DELETE("/:uid", s.recoverAttack)
	}
}

func (s *httpServer) createProcessAttack(c *gin.Context) {
	attack := &core.ProcessCommand{}
	if err := c.ShouldBindJSON(attack); err != nil {
		c.AbortWithError(http.StatusBadRequest, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	uid, err := s.chaos.ProcessAttack(attack)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, utils.AttackSuccessResponse(uid))
}

func (s *httpServer) createStressAttack(c *gin.Context) {
	attack := &core.StressCommand{}
	if err := c.ShouldBindJSON(attack); err != nil {
		c.AbortWithError(http.StatusBadRequest, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	uid, err := s.chaos.StressAttack(attack)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, utils.AttackSuccessResponse(uid))
}

func (s *httpServer) recoverAttack(c *gin.Context) {
	uid := c.Param("uid")
	err := utils.RecoverExp(s.exp, s.chaos, uid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, utils.RecoverSuccessResponse(uid))
}
