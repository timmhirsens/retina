// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package hubble

import (
	"context"
	"fmt"
	"path/filepath"

	"os"
	"sync"

	operatorOption "github.com/cilium/cilium/operator/option"
	"github.com/cilium/cilium/pkg/components"
	"github.com/cilium/cilium/pkg/hive"
	k8sClient "github.com/cilium/cilium/pkg/k8s/client"
	k8sversion "github.com/cilium/cilium/pkg/k8s/version"
	"github.com/cilium/cilium/pkg/logging"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/metrics"
	"github.com/cilium/cilium/pkg/option"
	"github.com/cilium/cilium/pkg/rand"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

var (
	// set at build time in Dockerfile
	applicationInsightsID string
	retinaVersion         string

	// set logger field: subsys=retina-operator
	binaryName = filepath.Base(os.Args[0])
	logger     = logging.DefaultLogger.WithField(logfields.LogSubsys, binaryName)
)

func NewOperatorCmd(h *hive.Hive) *cobra.Command {
	cmd := &cobra.Command{
		Use:   binaryName,
		Short: "Run " + binaryName,
		Run: func(cobraCmd *cobra.Command, args []string) {
			cmdRefDir := h.Viper().GetString(option.CMDRef)
			if cmdRefDir != "" {
				genMarkdown(cobraCmd, cmdRefDir)
				os.Exit(0)
			}

			initEnv(h.Viper())

			if err := h.Run(); err != nil {
				logger.Fatal(err)
			}
		},
	}

	h.RegisterFlags(cmd.Flags())

	// Enable fallback to direct API probing to check for support of Leases in
	// case Discovery API fails.
	h.Viper().Set(option.K8sEnableAPIDiscovery, true)

	// Overwrite the metrics namespace with the one specific for the Operator
	metrics.Namespace = RetinaOperatorMetricsNamespace

	cmd.AddCommand(
		MetricsCmd,
		h.Command(),
	)

	InitGlobalFlags(cmd, h.Viper())
	// not sure where flags hooks is set
	for _, hook := range FlagsHooks {
		hook.RegisterProviderFlag(cmd, h.Viper())
	}

	// the Operator performs config.Populate() within initEnv() above
	// the Daemon instead performs config.Populate() here on cobra.OnInitialize()
	cobra.OnInitialize(
		option.InitConfig(cmd, "Retina-Operator", "retina-operators", h.Viper()),
	)

	return cmd
}

func Execute(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func registerOperatorHooks(l logrus.FieldLogger, lc hive.Lifecycle, llc *LeaderLifecycle, clientset k8sClient.Clientset, shutdowner hive.Shutdowner) {
	var wg sync.WaitGroup
	lc.Append(hive.Hook{
		OnStart: func(hive.HookContext) error {
			wg.Add(1)
			go func() {
				runOperator(l, llc, clientset, shutdowner)
				wg.Done()
			}()
			return nil
		},
		OnStop: func(ctx hive.HookContext) error {
			if err := llc.Stop(ctx); err != nil {
				return errors.Wrap(err, "failed to stop operator")
			}
			doCleanup()
			wg.Wait()
			return nil
		},
	})
}

func initEnv(vp *viper.Viper) {
	// Prepopulate option.Config with options from CLI.

	// NOTE: if the flag is not provided in operator/cmd/flags.go InitGlobalFlags(), these Populate methods override
	// the default values provided in option.Config or operatorOption.Config respectively.
	// The values will be overriden to the "zero value".
	// Maybe could create a cell.Config for these instead?
	option.Config.Populate(vp)
	operatorOption.Config.Populate(vp)

	// add hooks after setting up metrics in the option.Confog
	logging.DefaultLogger.Hooks.Add(metrics.NewLoggingHook(components.CiliumOperatortName))

	// Logging should always be bootstrapped first. Do not add any code above this!
	if err := logging.SetupLogging(option.Config.LogDriver, logging.LogOptions(option.Config.LogOpt), binaryName, option.Config.Debug); err != nil {
		logger.Fatal(err)
	}

	option.LogRegisteredOptions(vp, logger)
	logger.Infof("retina operator version: %s", retinaVersion)
}

func doCleanup() {
	isLeader.Store(false)

	// Cancelling this context here makes sure that if the operator hold the
	// leader lease, it will be released.
	leaderElectionCtxCancel()
}

// runOperator implements the logic of leader election for cilium-operator using
// built-in leader election capability in kubernetes.
// See: https://github.com/kubernetes/client-go/blob/master/examples/leader-election/main.go
func runOperator(l logrus.FieldLogger, lc *LeaderLifecycle, clientset k8sClient.Clientset, shutdowner hive.Shutdowner) {
	isLeader.Store(false)

	leaderElectionCtx, leaderElectionCtxCancel = context.WithCancel(context.Background())

	// We only support Operator in HA mode for Kubernetes Versions having support for
	// LeasesResourceLock.
	// See docs on capabilities.LeasesResourceLock for more context.
	if !k8sversion.Capabilities().LeasesResourceLock {
		l.Info("Support for coordination.k8s.io/v1 not present, fallback to non HA mode")

		if err := lc.Start(leaderElectionCtx); err != nil {
			l.WithError(err).Fatal("Failed to start leading")
		}
		return
	}

	// Get hostname for identity name of the lease lock holder.
	// We identify the leader of the operator cluster using hostname.
	operatorID, err := os.Hostname()
	if err != nil {
		l.WithError(err).Fatal("Failed to get hostname when generating lease lock identity")
	}
	operatorID = rand.RandomStringWithPrefix(operatorID+"-", 10)

	leResourceLock, err := resourcelock.NewFromKubeconfig(
		resourcelock.LeasesResourceLock,
		operatorK8sNamespace,
		leaderElectionResourceLockName,
		resourcelock.ResourceLockConfig{
			// Identity name of the lock holder
			Identity: operatorID,
		},
		clientset.RestConfig(),
		operatorOption.Config.LeaderElectionRenewDeadline)
	if err != nil {
		l.WithError(err).Fatal("Failed to create resource lock for leader election")
	}

	// Start the leader election for running cilium-operators
	l.Info("Waiting for leader election")
	leaderelection.RunOrDie(leaderElectionCtx, leaderelection.LeaderElectionConfig{
		Name: leaderElectionResourceLockName,

		Lock:            leResourceLock,
		ReleaseOnCancel: true,

		LeaseDuration: operatorOption.Config.LeaderElectionLeaseDuration,
		RenewDeadline: operatorOption.Config.LeaderElectionRenewDeadline,
		RetryPeriod:   operatorOption.Config.LeaderElectionRetryPeriod,

		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				if err := lc.Start(ctx); err != nil {
					l.WithError(err).Error("Failed to start when elected leader, shutting down")
					shutdowner.Shutdown(hive.ShutdownWithError(err))
				}
			},
			OnStoppedLeading: func() {
				l.WithField("operator-id", operatorID).Info("Leader election lost")
				// Cleanup everything here, and exit.
				shutdowner.Shutdown(hive.ShutdownWithError(errors.New("Leader election lost")))
			},
			OnNewLeader: func(identity string) {
				if identity == operatorID {
					l.Info("Leading the operator HA deployment")
				} else {
					l.WithFields(logrus.Fields{
						"newLeader":  identity,
						"operatorID": operatorID,
					}).Info("Leader re-election complete")
				}
			},
		},
	})
}
