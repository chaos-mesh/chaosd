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
	"github.com/joomcode/errorx"
	"github.com/pingcap/log"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/chaos-mesh/chaosd/pkg/config"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/scheduler"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/server/utils"
	"github.com/chaos-mesh/chaosd/pkg/swaggerserver"
)

type httpServer struct {
	conf  *config.Config
	chaos *chaosd.Server
	exp   core.ExperimentStore
}

func NewServer(
	conf *config.Config,
	chaos *chaosd.Server,
	exp core.ExperimentStore,
) *httpServer {
	return &httpServer{
		conf:  conf,
		chaos: chaos,
		exp:   exp,
	}
}

func Register(s *httpServer, scheduler scheduler.Scheduler) {
	if s.conf.Platform != config.LocalPlatform {
		return
	}

	go func() {
		if err := s.startHttpServer(); err != nil {
			log.Fatal("failed to start HTTP server", zap.Error(err))
		}
	}()
	go func() {
		if err := s.startHttpsServer(); err != nil {
			log.Fatal("failed to start HTTPS server", zap.Error(err))
		}
	}()
	scheduler.Start()
}

func (s *httpServer) startHttpServer() error {
	httpServerAddr := s.conf.Address()
	log.Info("starting HTTP server", zap.String("address", httpServerAddr))
	e := gin.Default()
	e.Use(utils.MWHandleErrors())
	s.systemHandler(e)
	if s.serverMode() == HTTPServer {
		s.handler(e)
	}
	return e.Run(httpServerAddr)
}

func (s *httpServer) handler(engine *gin.Engine) {
	api := engine.Group("/api")
	{
		api.GET("/swagger/*any", swaggerserver.Handler())
	}

	attack := api.Group("/attack")
	{
		attack.POST("/process", s.createProcessAttack)
		attack.POST("/stress", s.createStressAttack)
		attack.POST("/network", s.createNetworkAttack)
		attack.POST("/disk", s.createDiskAttack)
		attack.POST("/clock", s.createClockAttack)
		attack.POST("/jvm", s.createJVMAttack)
		attack.POST("/redis", s.createRedisAttack)

		attack.DELETE("/:uid", s.recoverAttack)
	}

	experiments := api.Group("/experiments")
	{
		experiments.GET("/", s.listExperiments)
		experiments.GET("/:uid/runs", s.listExperimentRuns)
	}
}

func (s *httpServer) systemHandler(engine *gin.Engine) {
	api := engine.Group("/api")
	system := api.Group("/system")
	{
		system.GET("/health", s.healthcheck)
		system.GET("/version", s.version)
	}
}

// @Summary Create process attack.
// @Description Create process attack.
// @Tags attack
// @Produce json
// @Param request body core.ProcessCommand true "Request body"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /api/attack/process [post]
func (s *httpServer) createProcessAttack(c *gin.Context) {
	attack := core.NewProcessCommand()
	if err := c.ShouldBindJSON(attack); err != nil {
		c.AbortWithError(http.StatusBadRequest, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	attack.CompleteDefaults()
	if err := attack.Validate(); err != nil {
		err = core.ErrAttackConfigValidation.Wrap(err, "attack config validation failed")
		handleError(c, err)
		return
	}

	uid, err := s.chaos.ExecuteAttack(chaosd.ProcessAttack, attack, core.ServerMode)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, utils.AttackSuccessResponse(uid))
}

// @Summary Create network attack.
// @Description Create network attack.
// @Tags attack
// @Produce json
// @Param request body core.NetworkCommand true "Request body"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /api/attack/network [post]
func (s *httpServer) createNetworkAttack(c *gin.Context) {
	attack := core.NewNetworkCommand()
	if err := c.ShouldBindJSON(attack); err != nil {
		c.AbortWithError(http.StatusBadRequest, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	attack.CompleteDefaults()
	if err := attack.Validate(); err != nil {
		err = core.ErrAttackConfigValidation.Wrap(err, "attack config validation failed")
		handleError(c, err)
		return
	}

	uid, err := s.chaos.ExecuteAttack(chaosd.NetworkAttack, attack, core.ServerMode)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, utils.AttackSuccessResponse(uid))
}

// @Summary Create stress attack.
// @Description Create stress attack.
// @Tags attack
// @Produce json
// @Param request body core.StressCommand true "Request body"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /api/attack/stress [post]
func (s *httpServer) createStressAttack(c *gin.Context) {
	attack := core.NewStressCommand()
	if err := c.ShouldBindJSON(attack); err != nil {
		c.AbortWithError(http.StatusBadRequest, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	attack.CompleteDefaults()
	if err := attack.Validate(); err != nil {
		err = core.ErrAttackConfigValidation.Wrap(err, "attack config validation failed")
		handleError(c, err)
		return
	}

	uid, err := s.chaos.ExecuteAttack(chaosd.StressAttack, attack, core.ServerMode)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, utils.AttackSuccessResponse(uid))
}

// @Summary Create disk attack.
// @Description Create disk attack.
// @Tags attack
// @Produce json
// @Param request body core.DiskOption true "Request body"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /api/attack/disk [post]
func (s *httpServer) createDiskAttack(c *gin.Context) {
	options := core.NewDiskOption()
	if err := c.ShouldBindJSON(options); err != nil {
		c.AbortWithError(http.StatusBadRequest, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	options.CompleteDefaults()
	attackConfig, err := options.PreProcess()
	if err != nil {
		err = core.ErrAttackConfigValidation.Wrap(err, "attack config validation failed")
		handleError(c, err)
		return
	}

	uid, err := s.chaos.ExecuteAttack(chaosd.DiskAttack, attackConfig, core.ServerMode)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, utils.AttackSuccessResponse(uid))
}

// @Summary Create clock attack.
// @Description Create clock attack.
// @Tags attack
// @Produce json
// @Param request body core.ClockOption true "Request body"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /api/attack/clock [post]
func (s *httpServer) createClockAttack(c *gin.Context) {
	options := core.NewClockOption()
	if err := c.ShouldBindJSON(options); err != nil {
		c.AbortWithError(http.StatusBadRequest, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	options.CompleteDefaults()
	err := options.PreProcess()
	if err != nil {
		err = core.ErrAttackConfigValidation.Wrap(err, "attack config validation failed")
		handleError(c, err)
		return
	}
	uid, err := s.chaos.ExecuteAttack(chaosd.ClockAttack, options, core.ServerMode)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, utils.AttackSuccessResponse(uid))
}

// @Summary Create JVM attack.
// @Description Create JVM attack.
// @Tags attack
// @Produce json
// @Param request body core.JVMCommand true "Request body"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /api/attack/jvm [post]
func (s *httpServer) createJVMAttack(c *gin.Context) {
	options := core.NewJVMCommand()
	if err := c.ShouldBindJSON(options); err != nil {
		c.AbortWithError(http.StatusBadRequest, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	options.CompleteDefaults()
	if err := options.Validate(); err != nil {
		err = core.ErrAttackConfigValidation.Wrap(err, "attack config validation failed")
		handleError(c, err)
		return
	}

	uid, err := s.chaos.ExecuteAttack(chaosd.JVMAttack, options, core.ServerMode)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, utils.AttackSuccessResponse(uid))
}

// @Summary Create redis attack.
// @Description Create redis attack.
// @Tags attack
// @Produce json
// @Param request body core.RedisCommand true "Request body"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /api/attack/redis [post]
func (s *httpServer) createRedisAttack(c *gin.Context) {
	attack := core.NewRedisCommand()
	if err := c.ShouldBindJSON(attack); err != nil {
		c.AbortWithError(http.StatusBadRequest, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	attack.CompleteDefaults()
	if err := attack.Validate(); err != nil {
		err = core.ErrAttackConfigValidation.Wrap(err, "attack config validation failed")
		handleError(c, err)
		return
	}

	uid, err := s.chaos.ExecuteAttack(chaosd.RedisAttack, attack, core.ServerMode)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, utils.AttackSuccessResponse(uid))
}

// @Summary Create recover attack.
// @Description Create recover attack.
// @Tags attack
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} utils.Response
// @Failure 500 {object} utils.APIError
// @Router /api/attack/{uid} [delete]
func (s *httpServer) recoverAttack(c *gin.Context) {
	uid := c.Param("uid")
	err := s.chaos.RecoverAttack(uid)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, utils.RecoverSuccessResponse(uid))
}

func handleError(c *gin.Context, err error) {
	if err == gorm.ErrRecordNotFound {
		_ = c.AbortWithError(http.StatusNotFound, utils.ErrNotFound.WrapWithNoMessage(err))
		return
	}
	if errorx.IsOfType(err, core.ErrAttackConfigValidation) {
		_ = c.AbortWithError(http.StatusBadRequest, utils.ErrInvalidRequest.WrapWithNoMessage(err))
	} else {
		_ = c.AbortWithError(http.StatusInternalServerError, utils.ErrInternalServer.WrapWithNoMessage(err))
	}
}
