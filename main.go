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

  var gracefulStop = make(chan os.Signal)
  signal.Notify(gracefulStop, os.Interrupt)

  srv := CreateHttpServer()
  go StartServer(srv)
  ShutdownServer(gracefulStop, srv)
}

func ShutdownServer(haltSignal <-chan os.Signal, srv *http.Server) {
  <-haltSignal
  log.Info().Msg("Shutting down server...")
  // shut down gracefully, but wait no longer than 30 seconds before halting
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
  defer cancel()
	srv.Shutdown(ctx)
	log.Info().Msg("Server gracefully stopped")
}

func StartServer(srv *http.Server) {
  log.Fatal().Err(srv.ListenAndServe())
}

func CreateHttpServer() *http.Server {
  mux := http.NewServeMux()
  mux.Handle("/ishealthy", http.HandlerFunc(HandleHealthCheck))
  return &http.Server{Addr: ":44444", Handler: mux}
}

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
  cli, err := client.NewEnvClient()
  defer cli.Close()
  if err != nil {
    log.Error().Err(err).Msg("Unable to create docker client")
    http.Error(w, "", http.StatusInternalServerError)
    return
  }

  status, err := IsNodeHealthy(cli)
  if err != nil {
    http.Error(w, "", http.StatusInternalServerError)
    return
  }

  if status {
    w.WriteHeader(http.StatusNoContent)
  } else {
    http.Error(w, "Node is not ready", http.StatusServiceUnavailable)
  }
}

func IsNodeHealthy(cli client.SystemAPIClient) (bool, error) {
  info, err := cli.Info(context.Background())
  if err != nil {
    log.Error().Err(err).Msg("Unable to get docker information, is docker running?")
    return false, err
  }

  if len(info.Swarm.NodeID) != 0 && len(info.Swarm.Cluster.ID) != 0 && info.Swarm.LocalNodeState == swarm.LocalNodeStateActive {
    log.Info().Msgf("Swarm ID %s, Node ID %s, Swarm State %s", info.Swarm.Cluster.ID, info.Swarm.NodeID, info.Swarm.LocalNodeState)
    return true, nil
  }
  log.Info().Msgf("Swarm State %s", info.Swarm.LocalNodeState)
  return false, nil
}
