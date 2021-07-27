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

package utils

import (
	"github.com/pingcap/log"
	"go.uber.org/zap"
)

// Edge represents an edge in graph
type Edge struct {
	Source uint32
	Target uint32
	Next   *Edge
}

// Graph represents a graph with link list
type Graph struct {
	head map[uint32]*Edge
}

// NewGraph creates a Graph
func NewGraph() *Graph {
	return &Graph{
		head: make(map[uint32]*Edge),
	}
}

// Insert inserts an Edge into a graph from source to target
func (g *Graph) Insert(source uint32, target uint32) {
	newEdge := Edge{
		Source: source,
		Target: target,
		Next:   g.head[source],
	}
	g.head[source] = &newEdge
}

// IterFrom starts iterating from source node
func (g *Graph) IterFrom(source uint32) *Edge {
	return g.head[source]
}

// Flatten flattens the subtree from source (without checking whether it's a tree)
func (g *Graph) Flatten(source uint32) []uint32 {
	current := g.head[source]

	var flatTree []uint32
	for {
		if current == nil {
			break
		}

		children := g.Flatten(current.Target)
		flatTree = append(flatTree, current.Target)
		flatTree = append(flatTree, children...)

		current = current.Next
	}

	log.Debug("get flatTree", zap.Uint32("source", source), zap.Uint32s("flatTree", flatTree))
	return flatTree
}
