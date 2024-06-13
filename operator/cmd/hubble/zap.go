// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package hubble

import (
	"io"

	zaphook "github.com/Sytten/logrus-zap-hook"
	"github.com/cilium/cilium/pkg/hive/cell"
	"github.com/cilium/cilium/pkg/logging"
	"github.com/cilium/cilium/pkg/option"
	"github.com/microsoft/retina/pkg/log"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"

	"github.com/microsoft/retina/operator/v2/config"
)

// TODO refactor to another package? Like shared/telemetry/

const logFileName = "retina-operator.log"

type params struct {
	cell.In

	Logger      logrus.FieldLogger
	K8sCfg      *rest.Config
	DaemonCfg   *option.DaemonConfig
	OperatorCfg config.Config
}

func setupZapHook(p params) {
	// modify default logger
	// properly report the caller (otherwise, will get caller=zap.go every time)
	logging.DefaultLogger.ReportCaller = true
	// discard default logger output in favor of zap
	logging.DefaultLogger.SetOutput(io.Discard)

	lOpts := &log.LogOpts{
		Level:                 p.DaemonCfg.LogOpt[logging.LevelOpt],
		File:                  false,
		FileName:              logFileName,
		MaxFileSizeMB:         100,
		MaxBackups:            3,
		MaxAgeDays:            30,
		ApplicationInsightsID: applicationInsightsID,
		EnableTelemetry:       p.OperatorCfg.EnableTelemetry,
	}

	persistentFields := []zap.Field{
		zap.String("version", retinaVersion),
		zap.String("apiserver", p.K8sCfg.Host),
	}

	log.SetupZapLogger(lOpts, persistentFields...)

	namedLogger := log.Logger().Named("retina-operator-v2")
	namedLogger.Info("Traces telemetry initialized with zapai", zap.String("version", retinaVersion), zap.String("appInsightsID", lOpts.ApplicationInsightsID))

	zapHook, err := zaphook.NewZapHook(namedLogger.Logger)
	if err != nil {
		p.Logger.WithError(err).Error("failed to create zap hook")
		return
	}

	logging.DefaultLogger.Hooks.Add(zapHook)
}