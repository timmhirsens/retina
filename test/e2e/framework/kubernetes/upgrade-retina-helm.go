// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package kubernetes

import (
	"fmt"
	"log"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
)

type UpgradeRetinaHelmChart struct {
	Namespace          string
	ReleaseName        string
	KubeConfigFilePath string
	ChartPath          string
	TagEnv             string
}

func (u *UpgradeRetinaHelmChart) Run() error {
	settings := cli.New()
	settings.KubeConfig = u.KubeConfigFilePath
	actionConfig := new(action.Configuration)

	err := actionConfig.Init(settings.RESTClientGetter(), u.Namespace, os.Getenv("HELM_DRIVER"), log.Printf)
	if err != nil {
		return fmt.Errorf("failed to initialize helm action config: %w", err)
	}

	client := action.NewUpgrade(actionConfig)

	chart, err := loader.Load(u.ChartPath)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
	}
	// enable pod level
	options := values.Options{
		ValueFiles: []string{"../../../profiles/localctx/values.yaml"},
	}
	provider := getter.All(settings)
	values, err := options.MergeValues(provider)
	if err != nil {
		return fmt.Errorf("failed to merge values: %w", err)
	}

	// upgrade chart
	rel, err := client.Run(u.ReleaseName, chart, values)
	if err != nil {
		PrintPodLogs(u.KubeConfigFilePath, u.Namespace, "k8s-app=retina")
		return fmt.Errorf("failed to upgrade chart: %w", err)
	}

	log.Printf("upgraded chart from path: %s in namespace: %s\n", rel.Name, rel.Namespace)
	// this will confirm the values set during installation
	log.Printf("chart values: %v\n", rel.Config)

	return nil
}

func (u *UpgradeRetinaHelmChart) Prevalidate() error {
	return nil
}

func (u *UpgradeRetinaHelmChart) Stop() error {
	return nil
}
