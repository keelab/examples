// Command topology-rollout demonstrates immutable topology epochs, weighted
// traffic handoff and call leases.
package main

import (
	"context"
	"fmt"

	"github.com/keelab/keelith/programmable/topology"
)

func main() {
	manager := topology.NewManager()
	first, err := topology.Activate(topology.Plan{
		Epoch:      1,
		Placements: map[topology.PlacementID]struct{}{"blue": {}},
		Components: map[topology.ComponentID]topology.PlacementID{
			"api":   "blue",
			"order": "blue",
		},
		Dependencies: map[topology.ComponentID]map[topology.ComponentID]topology.BindingMode{
			"api": {"order": topology.BindingAuto},
		},
	})
	if err != nil {
		panic(fmt.Errorf("activate epoch 1: %w", err))
	}
	if err := manager.Stage(first); err != nil {
		panic(fmt.Errorf("stage epoch 1: %w", err))
	}
	if err := manager.Ready(first.Epoch()); err != nil {
		panic(fmt.Errorf("ready epoch 1: %w", err))
	}

	lease, err := manager.AcquireKey("order-001")
	if err != nil {
		panic(fmt.Errorf("acquire epoch 1 lease: %w", err))
	}
	fmt.Printf("before rollout: epoch=%d\n", lease.Epoch())
	lease.Release()

	second, err := topology.Activate(topology.Plan{
		Epoch:      2,
		Placements: map[topology.PlacementID]struct{}{"green": {}},
		Components: map[topology.ComponentID]topology.PlacementID{
			"api":   "green",
			"order": "green",
		},
		Dependencies: map[topology.ComponentID]map[topology.ComponentID]topology.BindingMode{
			"api": {"order": topology.BindingAuto},
		},
		Traffic: []topology.EpochWeight{
			{Epoch: 1, BasisPoints: 0},
			{Epoch: 2, BasisPoints: topology.TotalBasisPoints},
		},
	})
	if err != nil {
		panic(fmt.Errorf("activate epoch 2: %w", err))
	}
	if err := manager.Stage(second); err != nil {
		panic(fmt.Errorf("stage epoch 2: %w", err))
	}
	if err := manager.Ready(second.Epoch()); err != nil {
		panic(fmt.Errorf("ready epoch 2: %w", err))
	}
	if err := manager.Drain(first.Epoch()); err != nil {
		panic(fmt.Errorf("drain epoch 1: %w", err))
	}
	if err := manager.Stop(context.Background(), first.Epoch()); err != nil {
		panic(fmt.Errorf("stop epoch 1: %w", err))
	}
	binding, err := second.Resolve("api", "order")
	if err != nil {
		panic(fmt.Errorf("resolve dependency: %w", err))
	}
	fmt.Printf("after rollout: epoch=%d mode=%s placement=%s\n",
		binding.Epoch,
		binding.Mode,
		binding.TargetPlacement,
	)
}
