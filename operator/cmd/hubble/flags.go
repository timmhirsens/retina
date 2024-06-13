// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package hubble

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	operatorOption "github.com/cilium/cilium/operator/option"
	"github.com/cilium/cilium/pkg/defaults"
	"github.com/cilium/cilium/pkg/option"
)

func InitGlobalFlags(cmd *cobra.Command, vp *viper.Viper) {
	flags := cmd.Flags()

	// include this line so that we don't see the following log from Cilium:
	// "Running Cilium with \"kvstore\"=\"\" requires identity allocation via CRDs. Changing identity-allocation-mode to \"crd\""
	flags.String(option.IdentityAllocationMode, option.IdentityAllocationModeCRD, "Identity allocation mode")

	flags.String(option.ConfigFile, "", `Configuration file (to configure the operator, this argument is required)`)
	option.BindEnv(vp, option.ConfigFile)

	flags.String(option.ConfigDir, "", `Configuration directory that contains a file for each option`)
	option.BindEnv(vp, option.ConfigDir)

	flags.BoolP(option.DebugArg, "D", false, "Enable debugging mode")
	option.BindEnv(vp, option.DebugArg)

	// NOTE: without this the option gets overriden from the default value to the zero value via option.Config.Populate(vp)
	// specifically, here options.Config.AllocatorListTimeout gets overriden from the default value to 0s
	flags.Duration(option.AllocatorListTimeoutName, defaults.AllocatorListTimeout, "timeout to list initial allocator state")
	// similar overriding happens for option.Config.KVstoreConnectivityTimeout
	flags.Duration(option.KVstoreConnectivityTimeout, defaults.KVstoreConnectivityTimeout, "Time after which an incomplete kvstore operation  is considered failed")
	// similar overriding happens for option.Config.KVstorePeriodicSync
	flags.Duration(option.KVstorePeriodicSync, defaults.KVstorePeriodicSync, "Periodic KVstore synchronization interval")

	flags.Duration(operatorOption.EndpointGCInterval, operatorOption.EndpointGCIntervalDefault, "GC interval for cilium endpoints")
	option.BindEnv(vp, operatorOption.EndpointGCInterval)

	flags.Bool(operatorOption.EnableMetrics, false, "Enable Prometheus metrics")
	option.BindEnv(vp, operatorOption.EnableMetrics)

	flags.StringSlice(option.LogDriver, []string{}, "Logging endpoints to use for example syslog")
	option.BindEnv(vp, option.LogDriver)

	flags.Var(option.NewNamedMapOptions(option.LogOpt, &option.Config.LogOpt, nil),
		option.LogOpt, `Log driver options for cilium-operator, `+
			`configmap example for syslog driver: {"syslog.level":"info","syslog.facility":"local4"}`)
	option.BindEnv(vp, option.LogOpt)

	flags.Bool(option.Version, false, "Print version information")
	option.BindEnv(vp, option.Version)

	flags.String(option.CMDRef, "", "Path to cmdref output directory")
	flags.MarkHidden(option.CMDRef)
	option.BindEnv(vp, option.CMDRef)

	flags.Duration(operatorOption.LeaderElectionLeaseDuration, 15*time.Second,
		"Duration that non-leader operator candidates will wait before forcing to acquire leadership")
	option.BindEnv(vp, operatorOption.LeaderElectionLeaseDuration)

	flags.Duration(operatorOption.LeaderElectionRenewDeadline, 10*time.Second,
		"Duration that current acting master will retry refreshing leadership in before giving up the lock")
	option.BindEnv(vp, operatorOption.LeaderElectionRenewDeadline)

	flags.Duration(operatorOption.LeaderElectionRetryPeriod, 2*time.Second,
		"Duration that LeaderElector clients should wait between retries of the actions")
	option.BindEnv(vp, operatorOption.LeaderElectionRetryPeriod)

	flags.Bool(option.EnableCiliumEndpointSlice, false, "If set to true, the CiliumEndpointSlice feature is enabled. If any CiliumEndpoints resources are created, updated, or deleted in the cluster, all those changes are broadcast as CiliumEndpointSlice updates to all of the Cilium agents.")
	option.BindEnv(vp, option.EnableCiliumEndpointSlice)

	flags.Duration(option.KVstoreLeaseTTL, defaults.KVstoreLeaseTTL, "Time-to-live for the KVstore lease.")
	flags.MarkHidden(option.KVstoreLeaseTTL)
	option.BindEnv(vp, option.KVstoreLeaseTTL)

	vp.BindPFlags(flags)
}

const (
	// pprofOperator enables pprof debugging endpoint for the operator
	pprofOperator = "operator-pprof"

	// pprofAddress is the port that the pprof listens on
	pprofAddress = "operator-pprof-address"

	// pprofPort is the port that the pprof listens on
	pprofPort = "operator-pprof-port"
)

// operatorPprofConfig holds the configuration for the operator pprof cell.
// Differently from the agent and the clustermesh-apiserver, the operator prefixes
// the pprof related flags with the string "operator-".
// To reuse the same cell, we need a different config type to map the same fields
// to the operator-specific pprof flag names.
type operatorPprofConfig struct {
	OperatorPprof        bool
	OperatorPprofAddress string
	OperatorPprofPort    uint16
}

func (def operatorPprofConfig) Flags(flags *pflag.FlagSet) {
	flags.Bool(pprofOperator, def.OperatorPprof, "Enable serving pprof debugging API")
	flags.String(pprofAddress, def.OperatorPprofAddress, "Address that pprof listens on")
	flags.Uint16(pprofPort, def.OperatorPprofPort, "Port that pprof listens on")
}