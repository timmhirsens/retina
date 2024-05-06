// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package dns

import (
	"time"

	"github.com/microsoft/retina/test/e2e/framework/kubernetes"
	"github.com/microsoft/retina/test/e2e/framework/types"
)

const sleepDelay = 5 * time.Second

func ValidateDNSMetrics() *types.Scenario {
	name := "DNS Metrics"
	steps := []*types.StepWrapper{
		{
			Step: &kubernetes.CreateAgnhostStatefulSet{
				AgnhostName:      "agnhost-a",
				AgnhostNamespace: "kube-system",
			},
		},
		{
			Step: &kubernetes.ExecInPod{
				PodName:      "agnhost-a-0",
				PodNamespace: "kube-system",
				Command:      "nslookup kubernetes.default",
			},
			Opts: &types.StepOptions{
				ExpectError:               false,
				SkipSavingParamatersToJob: true,
			},
		},
		{
			Step: &types.Sleep{
				Duration: sleepDelay,
			},
		},
		{
			Step: &kubernetes.PortForward{
				Namespace:             "kube-system",
				LabelSelector:         "k8s-app=retina",
				LocalPort:             "10093",
				RemotePort:            "10093",
				OptionalLabelAffinity: "app=agnhost-a", // port forward to a pod on a node that also has this pod with this label, assuming same namespace
			},
			Opts: &types.StepOptions{
				RunInBackgroundWithID: "dns-port-forward",
			},
		},
		{
			Step: &ValidateDNSRequestMetrics{},
		},
		{
			Step: &ValidateDNSResponseMetrics{},
		},
		{
			Step: &types.Stop{
				BackgroundID: "dns-port-forward",
			},
		},
	}
	return types.NewScenario(name, steps...)
}
