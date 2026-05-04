package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/koldun11/yartime/client/internal/api"
	"github.com/koldun11/yartime/client/internal/enforce"
	"github.com/koldun11/yartime/client/internal/loop"
	"github.com/koldun11/yartime/client/internal/store/filestore"
	"go.uber.org/zap"
)

var version = "0.0.0"

func main() {
	serverURL := flag.String("server-url", "", "Yartime server URL (required)")
	clientID := flag.String("client-id", "default-client", "Client identifier")
	interval := flag.Duration("interval", 3*time.Minute, "Config poll interval")
	flag.Parse()

	if *serverURL == "" {
		fmt.Fprintln(os.Stderr, "error: --server-url is required")
		os.Exit(1)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting yartime client", zap.String("version", version))

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Fatal("failed to get home dir", zap.Error(err))
	}
	dbPath := filepath.Join(home, ".yartime")

	st, err := filestore.Open(dbPath)
	if err != nil {
		logger.Fatal("failed to open store", zap.Error(err))
	}
	defer st.Close()

	cfg, err := st.GetConfig()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}
	cfg.ServerURL = *serverURL
	cfg.ClientID = *clientID

	apiClient := api.New(*serverURL)

	newCfg, err := apiClient.FetchConfig(*clientID)
	if err != nil {
		logger.Warn("Failed to fetch initial config from server, using local", zap.Error(err))
	} else {
		newCfg.ServerURL = *serverURL
		if err := st.SaveConfig(newCfg); err != nil {
			logger.Error("Failed to save initial config", zap.Error(err))
		} else {
			cfg = newCfg
			logger.Info("Initial config fetched from server")
		}
	}

	if cfg.ExecuteOnStart != "" {
		logger.Info("ExecuteOnStart is set but not implemented yet")
	}

	enforcer := enforce.NewLinuxEnforcer(logger)

	binaryPath, err := os.Executable()
	if err != nil {
		logger.Warn("Failed to determine executable path", zap.Error(err))
		binaryPath = os.Args[0]
	}

	stop := make(chan struct{})
	go func() {
		loop.Run(cfg, st, apiClient, enforcer, *interval, version, binaryPath, logger, stop)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))

	close(stop)
	time.Sleep(500 * time.Millisecond)
}
