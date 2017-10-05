package main

import (
  "os"
  "net/http"
  "github.com/rs/zerolog"
  "github.com/rs/zerolog/log"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

func main() {
  zerolog.SetGlobalLevel(zerolog.InfoLevel)
  cli, err := client.NewEnvClient()
  if err != nil {
    log.Error().Err(err).Msg("Unable to create docker client")
    os.Exit(1)
  }

  http.HandleFunc("/ishealthy", func(w http.ResponseWriter, r *http.Request) {
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
  })
  log.Fatal().Err(http.ListenAndServe(":44444", nil))
  cli.Close()
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
