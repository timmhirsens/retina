// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Retina and Cilium

// NOTE: we could reference this file Cilium's code, but it is a small file.
// If we were to import this code from Cilium's operator/cmd/ package,
// that would require dependencies we don't need from their operator (for instance, BGP dependencies).
// At time of writing, trying to import that code was also resulting in an error in go mod tidy:
// module go.universe.tf/metallb@latest found (v0.13.12), but does not contain package go.universe.tf/metallb/pkg/speaker

package cmd

import (
	"github.com/cilium/cilium/pkg/hive"
	"github.com/cilium/cilium/pkg/hive/cell"
)

// LeaderLifecycle is the inner lifecycle of the operator that is started when this
// operator instance is elected leader. It implements hive.Lifecycle allowing cells
// to use it.
type LeaderLifecycle struct {
	hive.DefaultLifecycle
}

func WithLeaderLifecycle(cells ...cell.Cell) cell.Cell {
	return cell.Module(
		"leader-lifecycle",
		"Operator Leader Lifecycle",

		cell.Provide(
			func() *LeaderLifecycle { return &LeaderLifecycle{} },
		),
		cell.Decorate(
			func(lc *LeaderLifecycle) hive.Lifecycle {
				return lc
			},
			cells...,
		),
	)
}
