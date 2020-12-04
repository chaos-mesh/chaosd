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

package dbstore

import (
	"path"

	"go.uber.org/zap"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/pingcap/log"

	"github.com/chaos-mesh/chaosd/pkg/utils"
)

const dataFile = "chaosd.dat"

// DB defines a db storage.
type DB struct {
	*gorm.DB
}

// NewDBStore returns a new DB
func NewDBStore() (*DB, error) {
	gormDB, err := gorm.Open(sqlite.Open(path.Join(utils.GetProgramPath(), dataFile)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Error("failed to open DB", zap.Error(err))
		return nil, err
	}

	db := &DB{
		gormDB,
	}

	return db, nil
}

//func DryDBStore() (*DB, error) {
//	gormDB, err := gorm.Open(sqlite.Open(path.Join(utils.GetProgramPath(), dataFile)), &gorm.Config{
//		Logger: logger.Default.LogMode(logger.Silent),
//	})
//	if err != nil {
//		log.Error("failed to open DB", zap.Error(err))
//		return nil, err
//	}
//
//	db := &DB{
//		gormDB,
//	}
//
//	return db, nil
//}
