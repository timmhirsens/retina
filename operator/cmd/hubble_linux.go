// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

import (
	"fmt"

	"github.com/cilium/cilium/pkg/hive"
	"github.com/microsoft/retina/operator/cmd/hubble"
	"github.com/spf13/cobra"
)

var (
	h         = hive.New(hubble.Operator)
	hubbleCmd = &cobra.Command{
		Use:   "retina-operator",
		Short: "Start the Retina operator",
		Run: func(cobraCmd *cobra.Command, _ []string) {
			if v, _ := cobraCmd.Flags().GetBool("version"); v {
				fmt.Println("Starting Retina Operator")
			}
			hubble.NewOperatorCmd(h).Execute()
		},
	}
)

func init() {
	h.RegisterFlags(hubbleCmd.Flags())
	hubbleCmd.AddCommand(h.Command())

	hubble.InitGlobalFlags(hubbleCmd, h.Viper())

	rootCmd.AddCommand(hubbleCmd)
}
