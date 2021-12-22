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

package core

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/pingcap/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

type IPSetRuleStore interface {
	List(ctx context.Context) ([]*IPSetRule, error)
	Set(ctx context.Context, rule *IPSetRule) error
	FindByExperiment(ctx context.Context, experiment string) ([]*IPSetRule, error)
	DeleteByExperiment(ctx context.Context, experiment string) error
}

type IPSetRule struct {
	gorm.Model
	// The name of ipset
	Name string `gorm:"index:name" json:"name"`
	// The contents of ipset
	Cidrs string `json:"cidrs"`
	// Experiment represents the experiment which the rule belong to.
	Experiment string `gorm:"index:experiment" json:"experiment"`
}

type Cidr struct {
	gorm.Model
	Cidr string
}

// ChainDirection represents the direction of chain
type ChainDirection string

type IptablesRuleStore interface {
	List(ctx context.Context) ([]*IptablesRule, error)
	Set(ctx context.Context, rule *IptablesRule) error
	FindByExperiment(ctx context.Context, experiment string) ([]*IptablesRule, error)
	DeleteByExperiment(ctx context.Context, experiment string) error
}

type IptablesRule struct {
	gorm.Model
	// The name of iptables chain
	Name string `gorm:"index:name" json:"name"`
	// The name of related ipset
	IPSets string `json:"ipsets"`
	// The block direction of this iptables rule
	Direction string `json:"direction"`
	// Experiment represents the experiment which the rule belong to.
	Experiment string `gorm:"index:experiment" json:"experiment"`

	Protocol string `json:"protocol"`
}

func (i *IptablesRule) ToChain() *pb.Chain {
	ch := &pb.Chain{
		Name:      i.Name,
		Ipsets:    strings.Split(i.IPSets, ","),
		Direction: pb.Chain_Direction(pb.Chain_Direction_value[i.Direction]),
		Target:    "DROP",
	}

	return ch
}

type IptablesRuleList []*IptablesRule

func (l IptablesRuleList) ToChains() []*pb.Chain {
	chains := make([]*pb.Chain, 0)

	for _, rule := range l {
		chains = append(chains, rule.ToChain())
	}

	return chains
}

type TCRuleStore interface {
	List(ctx context.Context) ([]*TCRule, error)
	ListGroupDevice(ctx context.Context) (map[string][]*TCRule, error)
	Set(ctx context.Context, rule *TCRule) error
	FindByDevice(ctx context.Context, experiment string) ([]*TCRule, error)
	FindByExperiment(ctx context.Context, experiment string) ([]*TCRule, error)
	DeleteByExperiment(ctx context.Context, experiment string) error
}

type TCRule struct {
	gorm.Model
	Device string `json:"device"`
	// The type of traffic control
	Type string `json:"type"`
	TC   string `json:"tc"`
	// The name of target ipset
	IPSet string `json:"ipset,omitempty"`
	// Experiment represents the experiment which the rule belong to.
	Experiment string `gorm:"index:experiment" json:"experiment"`

	Protocal   string
	SourcePort string
	EgressPort string
}

func (t *TCRule) ToTC() (*pb.Tc, error) {
	tc := &pb.Tc{
		Ipset:      t.IPSet,
		Protocol:   t.Protocal,
		SourcePort: t.SourcePort,
		EgressPort: t.EgressPort,
	}

	tcp := &TcParameter{}
	if err := json.Unmarshal([]byte(t.TC), tcp); err != nil {
		return nil, errors.WithStack(err)
	}

	switch t.Type {
	case pb.Tc_BANDWIDTH.String():
		tbf, err := tcp.Bandwidth.ToTbf()
		if err != nil {
			return nil, errors.WithStack(err)
		}

		tc.Type = pb.Tc_BANDWIDTH
		tc.Tbf = tbf
	case pb.Tc_NETEM.String():
		netem, err := toNetem(tcp)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		tc.Type = pb.Tc_NETEM
		tc.Netem = netem
	}

	return tc, nil
}

type TCRuleList []*TCRule

func (t TCRuleList) ToTCs() ([]*pb.Tc, error) {
	tcs := make([]*pb.Tc, 0)
	for _, rule := range t {
		tc, err := rule.ToTC()
		if err != nil {
			return nil, errors.WithStack(err)
		}

		tcs = append(tcs, tc)
	}

	return tcs, nil
}

// TcParameter represents the parameters for a traffic control chaos
type TcParameter struct {
	Device string
	// Delay represents the detail about delay action
	Delay *DelaySpec `json:"delay,omitempty"`
	// Loss represents the detail about loss action
	Loss *LossSpec `json:"loss,omitempty"`
	// DuplicateSpec represents the detail about loss action
	Duplicate *DuplicateSpec `json:"duplicate,omitempty"`
	// Corrupt represents the detail about corrupt action
	Corrupt *CorruptSpec `json:"corrupt,omitempty"`
	// Bandwidth represents the detail about bandwidth control action
	Bandwidth *BandwidthSpec `json:"bandwidth,omitempty"`
}

// DelaySpec defines detail of a delay action
type DelaySpec struct {
	Latency     string       `json:"latency"`
	Correlation string       `json:"correlation,omitempty"`
	Jitter      string       `json:"jitter,omitempty"`
	Reorder     *ReorderSpec `json:"reorder,omitempty"`
}

// ToNetem implements Netem interface.
func (in *DelaySpec) ToNetem() (*pb.Netem, error) {
	delayTime, err := time.ParseDuration(in.Latency)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	jitter, err := time.ParseDuration(in.Jitter)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	corr, err := strconv.ParseFloat(in.Correlation, 32)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	netem := &pb.Netem{
		Time:      uint32(delayTime.Nanoseconds() / 1e3),
		DelayCorr: float32(corr),
		Jitter:    uint32(jitter.Nanoseconds() / 1e3),
	}

	if in.Reorder != nil {
		reorderPercentage, err := strconv.ParseFloat(in.Reorder.Reorder, 32)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		corr, err := strconv.ParseFloat(in.Reorder.Correlation, 32)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		netem.Reorder = float32(reorderPercentage)
		netem.ReorderCorr = float32(corr)
		netem.Gap = uint32(in.Reorder.Gap)
	}

	return netem, nil
}

// ReorderSpec defines details of packet reorder.
type ReorderSpec struct {
	Reorder     string `json:"reorder"`
	Correlation string `json:"correlation"`
	Gap         int    `json:"gap"`
}

// LossSpec defines detail of a loss action
type LossSpec struct {
	Loss        string `json:"loss"`
	Correlation string `json:"correlation"`
}

// ToNetem implements Netem interface.
func (in *LossSpec) ToNetem() (*pb.Netem, error) {
	lossPercentage, err := strconv.ParseFloat(in.Loss, 32)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	corr, err := strconv.ParseFloat(in.Correlation, 32)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &pb.Netem{
		Loss:     float32(lossPercentage),
		LossCorr: float32(corr),
	}, nil
}

// DuplicateSpec defines detail of a duplicate action
type DuplicateSpec struct {
	Duplicate   string `json:"duplicate"`
	Correlation string `json:"correlation"`
}

// ToNetem implements Netem interface.
func (in *DuplicateSpec) ToNetem() (*pb.Netem, error) {
	duplicatePercentage, err := strconv.ParseFloat(in.Duplicate, 32)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	corr, err := strconv.ParseFloat(in.Correlation, 32)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &pb.Netem{
		Duplicate:     float32(duplicatePercentage),
		DuplicateCorr: float32(corr),
	}, nil
}

// CorruptSpec defines detail of a corrupt action
type CorruptSpec struct {
	Corrupt     string `json:"corrupt"`
	Correlation string `json:"correlation"`
}

// ToNetem implements Netem interface.
func (in *CorruptSpec) ToNetem() (*pb.Netem, error) {
	corruptPercentage, err := strconv.ParseFloat(in.Corrupt, 32)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	corr, err := strconv.ParseFloat(in.Correlation, 32)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &pb.Netem{
		Corrupt:     float32(corruptPercentage),
		CorruptCorr: float32(corr),
	}, nil
}

// BandwidthSpec defines detail of bandwidth limit.
type BandwidthSpec struct {
	// Rate is the speed knob. Allows bps, kbps, mbps, gbps, tbps unit. bps means bytes per second.
	Rate string `json:"rate"`
	// Limit is the number of bytes that can be queued waiting for tokens to become available.
	Limit uint32 `json:"limit"`
	// Buffer is the maximum amount of bytes that tokens can be available for instantaneously.
	Buffer uint32 `json:"buffer"`
	// Peakrate is the maximum depletion rate of the bucket.
	// The peakrate does not need to be set, it is only necessary
	// if perfect millisecond timescale shaping is required.
	Peakrate *uint64 `json:"peakrate,omitempty"`
	// Minburst specifies the size of the peakrate bucket. For perfect
	// accuracy, should be set to the MTU of the interface.  If a
	// peakrate is needed, but some burstiness is acceptable, this
	// size can be raised. A 3000 byte minburst allows around 3mbit/s
	// of peakrate, given 1000 byte packets.
	Minburst *uint32 `json:"minburst,omitempty"`
}

// ToTbf converts BandwidthSpec to *chaosdaemonpb.Tbf
// Bandwidth action use TBF under the hood.
// TBF stands for Token Bucket Filter, is a classful queueing discipline available
// for traffic control with the tc command.
// http://man7.org/linux/man-pages/man8/tc-tbf.8.html
func (in *BandwidthSpec) ToTbf() (*pb.Tbf, error) {
	rate, err := convertUnitToBytes(in.Rate)

	if err != nil {
		return nil, err
	}

	tbf := &pb.Tbf{
		Rate:   rate,
		Limit:  in.Limit,
		Buffer: in.Buffer,
	}

	if in.Peakrate != nil && in.Minburst != nil {
		tbf.PeakRate = *in.Peakrate
		tbf.MinBurst = *in.Minburst
	}

	return tbf, nil
}

func convertUnitToBytes(nu string) (uint64, error) {
	// normalize input
	s := strings.ToLower(strings.TrimSpace(nu))

	for i, u := range []string{"tbps", "gbps", "mbps", "kbps", "bps"} {
		if strings.HasSuffix(s, u) {
			ts := strings.TrimSuffix(s, u)
			s := strings.TrimSpace(ts)

			n, err := strconv.ParseUint(s, 10, 64)

			if err != nil {
				return 0, err
			}

			// convert unit to bytes
			for j := 4 - i; j > 0; j-- {
				n = n * 1024
			}

			return n, nil
		}
	}

	return 0, errors.New("invalid unit")
}

// NetemSpec defines the interface to convert to a Netem protobuf
type NetemSpec interface {
	ToNetem() (*pb.Netem, error)
}

// toNetem calls ToNetem on all non nil network emulation specs and merges them into one request.
func toNetem(spec *TcParameter) (*pb.Netem, error) {
	// NOTE: a cleaner way like
	// emSpecs = []NetemSpec{spec.Delay, spec.Loss} won't work.
	// Because in the for _, spec := range emSpecs loop,
	// spec != nil would always be true.
	// See https://stackoverflow.com/questions/13476349/check-for-nil-and-nil-interface-in-go
	// And https://groups.google.com/forum/#!topic/golang-nuts/wnH302gBa4I/discussion
	// > In short: If you never store (*T)(nil) in an interface, then you can reliably use comparison against nil
	var emSpecs []NetemSpec
	if spec.Delay != nil {
		emSpecs = append(emSpecs, spec.Delay)
	}
	if spec.Loss != nil {
		emSpecs = append(emSpecs, spec.Loss)
	}
	if spec.Duplicate != nil {
		emSpecs = append(emSpecs, spec.Duplicate)
	}
	if spec.Corrupt != nil {
		emSpecs = append(emSpecs, spec.Corrupt)
	}
	if len(emSpecs) == 0 {
		return nil, errors.New("invalid netem")
	}

	merged := &pb.Netem{}
	for _, spec := range emSpecs {
		em, err := spec.ToNetem()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		merged = mergeNetem(merged, em)
	}
	return merged, nil
}

// mergeNetem merges two Netem protos into a new one.
// REMEMBER to assign the return value, i.e. merged = utils.MergeNetm(merged, em)
// For each field it takes the bigger value of the two.
// Its main use case is merging netem of different types, e.g. delay and loss.
// It returns nil if both inputs are nil.
// Otherwise it returns a new Netem with merged values.
func mergeNetem(a, b *pb.Netem) *pb.Netem {
	if a == nil && b == nil {
		return nil
	}
	// NOTE: because proto getters check nil, we are good here even if one of them is nil.
	// But we just assign empty value to make IDE and linters happy.
	if a == nil {
		a = &pb.Netem{}
	}
	if b == nil {
		b = &pb.Netem{}
	}
	return &pb.Netem{
		Time:          maxu32(a.GetTime(), b.GetTime()),
		Jitter:        maxu32(a.GetJitter(), b.GetJitter()),
		DelayCorr:     maxf32(a.GetDelayCorr(), b.GetDelayCorr()),
		Limit:         maxu32(a.GetLimit(), b.GetLimit()),
		Loss:          maxf32(a.GetLoss(), b.GetLoss()),
		LossCorr:      maxf32(a.GetLossCorr(), b.GetLossCorr()),
		Gap:           maxu32(a.GetGap(), b.GetGap()),
		Duplicate:     maxf32(a.GetDuplicate(), b.GetDuplicate()),
		DuplicateCorr: maxf32(a.GetDuplicateCorr(), b.GetDuplicateCorr()),
		Reorder:       maxf32(a.GetReorder(), b.GetReorder()),
		ReorderCorr:   maxf32(a.GetReorderCorr(), b.GetReorderCorr()),
		Corrupt:       maxf32(a.GetCorrupt(), b.GetCorrupt()),
		CorruptCorr:   maxf32(a.GetCorruptCorr(), b.GetCorruptCorr()),
	}
}

func maxu32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func maxf32(a, b float32) float32 {
	return float32(math.Max(float64(a), float64(b)))
}
