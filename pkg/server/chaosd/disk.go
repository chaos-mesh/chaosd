package chaosd

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/rand"
	"os"
	"os/exec"
	"strconv"

	"github.com/google/uuid"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

const DDWritePayloadCommand = "dd if=/dev/zero of=%s bs=%s count=%s oflag=dsync"
const DDReadPayloadCommand = "dd if=%s of=/dev/null bs=%s count=%s iflag=dsync,direct,fullblock"

func (s *Server) DiskPayload(fill *core.DiskCommand) (uid string, err error) {
	uid = uuid.New().String()

	if err = s.exp.Set(context.Background(), &core.Experiment{
		Uid:            uid,
		Status:         core.Created,
		Kind:           core.DiskAttack,
		Action:         fill.Action,
		RecoverCommand: fill.String(),
	}); err != nil {
		err = errors.WithStack(err)
		return
	}
	defer func() {
		if err != nil {
			if err := s.exp.Update(context.Background(), uid, core.Error, err.Error(), fill.String()); err != nil {
				log.Error("failed to update experiment", zap.Error(err))
			}
			return
		}
		if err := s.exp.Update(context.Background(), uid, core.Success, "", fill.String()); err != nil {
			log.Error("failed to update experiment", zap.Error(err))
		}
	}()

	switch fill.Action {
	case core.DiskWritePayloadAction:
		cmd := exec.Command(fmt.Sprintf(DDWritePayloadCommand, fill.Path, "1M", strconv.FormatUint(fill.Size, 10)))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(string(output), zap.Error(err))
		}
		return uid, err
	case DDReadPayloadCommand:
		cmd := exec.Command(fmt.Sprintf(DDReadPayloadCommand, fill.Path, "1M", strconv.FormatUint(fill.Size, 10)))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(string(output), zap.Error(err))
		}
		return uid, err
	default:
		err := errors.Errorf("invalid fill action")
		log.Error(fmt.Sprintf("fill action: %s", fill.Action), zap.Error(err))
		return uid, err
	}
}

const DDFillCommand = "dd if=/dev/zero of=%s bs=%s count=%s iflag=fullblock"

func (s *Server) DiskFill(fill *core.DiskCommand) (uid string, err error) {
	uid = uuid.New().String()

	if err = s.exp.Set(context.Background(), &core.Experiment{
		Uid:            uid,
		Status:         core.Created,
		Kind:           core.DiskAttack,
		Action:         fill.Action,
		RecoverCommand: fill.String(),
	}); err != nil {
		err = errors.WithStack(err)
		return
	}
	defer func() {
		if err != nil {
			if err := s.exp.Update(context.Background(), uid, core.Error, err.Error(), fill.String()); err != nil {
				log.Error("failed to update experiment", zap.Error(err))
			}
			return
		}
		if err := s.exp.Update(context.Background(), uid, core.Success, "", fill.String()); err != nil {
			log.Error("failed to update experiment", zap.Error(err))
		}
	}()

	if fill.Path == "" {
		for {
			tempFilename := rand.String(128) + "temp-data"
			if _, err := os.Stat(tempFilename); os.IsNotExist(err) {
				fill.Path = tempFilename
				_, err = os.Create(fill.Path)
				if err != nil {
					log.Error("unexpected err when creating temp file", zap.Error(err))
					return uid, err
				}
				break
			}
		}
		defer func() {
			err := os.Remove(fill.Path)
			if err != nil {
				log.Error(fmt.Sprintf("unexpected err when removing temp file %s", fill.Path), zap.Error(err))
			}
		}()

	}

	cmd := exec.Command(fmt.Sprintf(DDFillCommand, fill.Path, "1M", strconv.FormatUint(fill.Size, 10)))
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Error(string(output), zap.Error(err))
	}

	return uid, err
}
