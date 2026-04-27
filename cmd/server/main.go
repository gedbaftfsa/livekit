// Copyright 2023 LiveKit, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/livekit/livekit-server/pkg/config"
	"github.com/livekit/livekit-server/pkg/logger"
	"github.com/livekit/livekit-server/pkg/service"
	"github.com/livekit/livekit-server/version"
)

func init() {
	// Seed the random number generator for session ID generation and other uses.
	rand.Seed(time.Now().UnixNano())
}

func main() {
	app := &cli.App{
		Name:    "livekit-server",
		Usage:   "LiveKit SFU server",
		Version: version.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "path to LiveKit config file",
				EnvVars: []string{"LIVEKIT_CONFIG_FILE"},
			},
			&cli.StringFlag{
				Name:    "config-body",
				Usage:   "LiveKit config in YAML, read from stdin or passed in directly",
				EnvVars: []string{"LIVEKIT_CONFIG_BODY"},
			},
			&cli.StringFlag{
				Name:    "key-file",
				Usage:   "path to file that contains API keys/secrets",
				EnvVars: []string{"LIVEKIT_KEY_FILE"},
			},
			&cli.StringFlag{
				Name:    "keys",
				Usage:   "API keys/secrets pairs in the format of key: secret",
				EnvVars: []string{"LIVEKIT_KEYS"},
			},
			&cli.StringFlag{
				Name:    "node-ip",
				Usage:   "IP address of the current node, used to advertise to clients",
				EnvVars: []string{"NODE_IP"},
			},
			&cli.StringFlag{
				Name:    "bind",
				Usage:   "address to bind to (default: 0.0.0.0)",
				EnvVars: []string{"LIVEKIT_BIND"},
			},
			&cli.BoolFlag{
				Name:  "dev",
				Usage: "sets log-level to debug, and console formatter",
			},
		},
		Action: startServer,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startServer(c *cli.Context) error {
	conf, err := config.NewConfig(c.String("config"), c.String("config-body"), c)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	logger.InitFromConfig(&conf.Logging, conf.Development)

	server, err := service.InitializeServer(conf)
	if err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	// Handle SIGINT, SIGTERM, and SIGQUIT for graceful shutdown.
	// Note: removed SIGHUP here since we don't actually support config reload yet;
	// catching it was causing the server to shut down unexpectedly on terminal hang-up.
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-sigChan
		// Log which signal triggered the shutdown so it's easier to diagnose
		// unexpected restarts in production logs.
		logger.GetLogger().Infow("received signal, shutting down", "signal", sig)
		server.Stop(false)
	}()

	return server.Start()
}
