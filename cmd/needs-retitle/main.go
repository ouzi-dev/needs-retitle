/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ouzi-dev/needs-retitle/pkg/config"
	"github.com/ouzi-dev/needs-retitle/pkg/plugin"
	"github.com/ouzi-dev/needs-retitle/pkg/server"
	"github.com/ouzi-dev/needs-retitle/pkg/version"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/interrupts"

	"k8s.io/test-infra/pkg/flagutil"
	"k8s.io/test-infra/prow/config/secret"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	pluginsflagutil "k8s.io/test-infra/prow/flagutil/plugins"

	"k8s.io/test-infra/prow/pluginhelp/externalplugins"
)

type options struct {
	port int

	pluginConfig pluginsflagutil.PluginOptions
	dryRun       bool
	github       prowflagutil.GitHubOptions

	updatePeriod time.Duration

	webhookSecretFile string
}

func (o *options) Validate() error {
	for idx, group := range []flagutil.OptionGroup{&o.github} {
		if err := group.Validate(o.dryRun); err != nil {
			return fmt.Errorf("%d: %w", idx, err)
		}
	}

	return nil
}

func gatherOptions() options {
	o := options{}
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.IntVar(&o.port, "port", 8888, "Port to listen on.")
	fs.BoolVar(&o.dryRun, "dry-run", true, "Dry run for testing. Uses API tokens but does not mutate.")
	fs.DurationVar(&o.updatePeriod, "update-period", time.Hour*24, "Period duration for periodic scans of all PRs.")
	fs.StringVar(&o.webhookSecretFile, "hmac-secret-file", "/etc/webhook/hmac", "Path to the file containing the GitHub HMAC secret.")

	o.pluginConfig.PluginConfigPathDefault = "/etc/plugins/plugins.yaml"

	for _, group := range []flagutil.OptionGroup{&o.github, &o.pluginConfig} {
		group.AddFlags(fs)
	}
	fs.Parse(os.Args[1:])
	return o
}

func main() {
	o := gatherOptions()
	if err := o.Validate(); err != nil {
		logrus.Fatalf("Invalid options: %v", err)
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	log := logrus.StandardLogger().WithField("plugin", plugin.PluginName)

	log.Infof("Starting plugin %s %s", plugin.PluginName, version.GetVersion())

	if err := secret.Add(o.github.TokenPath, o.webhookSecretFile); err != nil {
		logrus.WithError(err).Fatal("Error starting secrets agent.")
	}

	pa, err := o.pluginConfig.PluginAgent()
	if err != nil {
		log.WithError(err).Fatal("Error loading plugin config")
	}

	pca := config.NewPluginConfigAgent()
	if err := pca.Start(o.pluginConfig.PluginConfigPath); err != nil {
		log.WithError(err).Fatalf("Error loading %s config from %q.", plugin.PluginName, o.pluginConfig)
	}

	githubClient, err := o.github.GitHubClient(o.dryRun)
	if err != nil {
		logrus.WithError(err).Fatal("Error getting GitHub client.")
	}
	githubClient.Throttle(360, 360)

	s := server.NewServer(secret.GetTokenGenerator(o.webhookSecretFile), githubClient, log, pca.GetPlugin())

	defer interrupts.WaitForGracefulShutdown()

	interrupts.TickLiteral(func() {
		start := time.Now()
		if err := pca.GetPlugin().HandleAll(log, githubClient, pa.Config()); err != nil {
			log.WithError(err).Error("Error during periodic update of all PRs.")
		}
		log.WithField("duration", fmt.Sprintf("%v", time.Since(start))).Info("Periodic update complete.")
	}, o.updatePeriod)

	mux := http.NewServeMux()
	mux.Handle("/", s)
	externalplugins.ServeExternalPluginHelp(mux, log, plugin.HelpProvider)
	httpServer := &http.Server{Addr: ":" + strconv.Itoa(o.port), Handler: mux}
	interrupts.ListenAndServe(httpServer, 5*time.Second)
}
