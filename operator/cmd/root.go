// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

import (
	"fmt"
	"os"

	"github.com/microsoft/retina/operator/cmd/legacy"
	"github.com/spf13/cobra"
)

var (
	metricsAddr          string
	probeAddr            string
	enableLeaderElection bool

	rootCmd = &cobra.Command{
		Use:   "retina-operator",
		Short: "Retina Operator",
		Long:  "Start Retina Operator",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
			fmt.Println("Starting Retina Operator")
			d := legacy.NewOperator(metricsAddr, probeAddr, enableLeaderElection)
			d.Start()
		},
	}
)

func init() {
	rootCmd.Flags().StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	rootCmd.Flags().StringVar(&probeAddr, "probe-addr", ":8081", "The address the probe endpoint binds to.")
	rootCmd.Flags().BoolVar(&enableLeaderElection, "enable-leader-election", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
