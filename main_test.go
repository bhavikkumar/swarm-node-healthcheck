package main

import (
  "testing"
  "context"
  "errors"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "github.com/docker/docker/client"
  "github.com/docker/docker/api/types"
  "github.com/docker/docker/api/types/events"
  "github.com/docker/docker/api/types/registry"
  "github.com/docker/docker/api/types/swarm"
  "github.com/satori/go.uuid"
)

func TestCannotGetDockerInfo(t *testing.T) {
  testMock := new(MockSystemAPIClient)
  testMock.On("Info", context.Background()).Return(types.Info{}, errors.New("Test - could not get info from docker"))
  healthy, err := IsNodeHealthy(testMock)

  assert.False(t, healthy)
  assert.Error(t, err)
}

func TestNodeNotHealthy(t *testing.T) {
  // Setup expected response from docker client
  var info types.Info
  info.Swarm.LocalNodeState = swarm.LocalNodeStateInactive

  testMock := new(MockSystemAPIClient)
  testMock.On("Info", context.Background()).Return(info, nil)
  healthy, err := IsNodeHealthy(testMock)

  assert.False(t, healthy)
  assert.Nil(t, err)
}

func TestNodeHealthy(t *testing.T) {
  // Setup expected response from docker client
  var info types.Info
  info.Swarm.NodeID = uuid.NewV4().String()
  info.Swarm.Cluster.ID = uuid.NewV4().String()
  info.Swarm.LocalNodeState = swarm.LocalNodeStateActive

  testMock := new(MockSystemAPIClient)
  testMock.On("Info", context.Background()).Return(info, nil)
  healthy, err := IsNodeHealthy(testMock)

  assert.True(t, healthy)
  assert.Nil(t, err)
}

func TestNodeWithoutClusterID(t *testing.T) {
  // Setup expected response from docker client
  var info types.Info
  info.Swarm.NodeID = uuid.NewV4().String()
  info.Swarm.LocalNodeState = swarm.LocalNodeStatePending

  testMock := new(MockSystemAPIClient)
  testMock.On("Info", context.Background()).Return(info, nil)
  healthy, err := IsNodeHealthy(testMock)

  assert.False(t, healthy)
  assert.Nil(t, err)
}

/*
  Test objects
*/

// DockerMock is a mocked object that implements an interface
// that describes an object that the code I am testing relies on.
type MockSystemAPIClient struct {
  mock.Mock
	client.SystemAPIClient
}

// Info is a method on MockSystemAPIClient that implements some interface
// and just records the activity, and returns what the Mock object tells it to.
// NOTE: This method is not being tested here, code that uses this object is
func (cli *MockSystemAPIClient) Info(ctx context.Context) (types.Info, error) {
  ret := cli.Called(ctx)
  return ret.Get(0).(types.Info), ret.Error(1)
}

func (cli *MockSystemAPIClient) Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
  return nil, nil
}

func (cli *MockSystemAPIClient) RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error) {
  return registry.AuthenticateOKBody{}, nil
}

func (cli *MockSystemAPIClient) DiskUsage(ctx context.Context) (types.DiskUsage, error) {
  return types.DiskUsage{}, nil
}

func (cli *MockSystemAPIClient) Ping(ctx context.Context) (types.Ping, error) {
  return types.Ping{}, nil
}
