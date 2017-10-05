package main

import (
  "context"
	"net/http"
  "os"
  "os/signal"
  "time"
  "github.com/rs/zerolog"
  "github.com/rs/zerolog/log"
  "github.com/docker/docker/api/types/swarm"
  "github.com/docker/docker/client"
)

func main() {
  zerolog.SetGlobalLevel(zerolog.InfoLevel)
  cli, err := client.NewEnvClient()
  if err != nil {
    log.Error().Err(err).Msg("Unable to create docker client")
    os.Exit(1)
  }

  // subscribe to SIGINT signals
  stopChan := make(chan os.Signal)
  signal.Notify(stopChan, os.Interrupt)

  mux := http.NewServeMux()

  mux.Handle("/ishealthy", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    status, err := isNodeHealthy(cli)
    if err != nil {
      http.Error(w, "", http.StatusInternalServerError)
      return
    }

    if status {
      w.WriteHeader(http.StatusNoContent)
    } else {
      http.Error(w, "Node is not ready", http.StatusServiceUnavailable)
    }
  }))

  srv := &http.Server{Addr: ":44444", Handler: mux}

  go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal().Err(err)
		}
	}()

	<-stopChan // wait for SIGINT
	log.Info().Msg("Shutting down server...")
	// shut down gracefully, but wait no longer than 5 seconds before halting
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
  cli.Close()
	srv.Shutdown(ctx)

	log.Info().Msg("Server gracefully stopped")
}

func isNodeHealthy(cli *client.Client) (bool, error) {
  info, err := cli.Info(context.Background())
  if err != nil {
    log.Error().Err(err).Msg("Unable to get docker information, is docker running?")
    return false, err
  }

  if info.Swarm.Cluster != nil && len(info.Swarm.NodeID) != 0 && len(info.Swarm.Cluster.ID) != 0 && info.Swarm.LocalNodeState == swarm.LocalNodeStateActive {
    log.Info().Msgf("Swarm ID %s, Node ID %s, Swarm State %s", info.Swarm.Cluster.ID, info.Swarm.NodeID, info.Swarm.LocalNodeState)
    return true, nil
  }
  log.Info().Msgf("Swarm State %s", info.Swarm.LocalNodeState)
  return false, nil
}
