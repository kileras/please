package cluster

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "cache/proto/rpc_cache"
	"cache/tools"
)

func TestBringUpCluster(t *testing.T) {
	c1 := newCluster(5995, 6995, "c1")
	m1 := newRPCServer(c1, 6995)
	c1.Init(3)
	log.Notice("Cluster seeded")

	c2 := newCluster(5996, 6996, "c2")
	m2 := newRPCServer(c2, 6996)
	c2.Join([]string{"127.0.0.1:5995"})
	log.Notice("c2 joined cluster")

	expected := []*pb.Node{
		&pb.Node{
			Name:      "c1",
			Address:   "127.0.0.1:6995",
			HashBegin: tools.HashPoint(0, 3),
			HashEnd:   tools.HashPoint(1, 3),
		},
		&pb.Node{
			Name:      "c2",
			Address:   "127.0.0.1:6996",
			HashBegin: tools.HashPoint(1, 3),
			HashEnd:   tools.HashPoint(2, 3),
		},
	}
	// Both nodes should agree about the member list
	assert.Equal(t, expected, c1.GetMembers())
	assert.Equal(t, expected, c2.GetMembers())

	c3 := newCluster(5997, 6997, "c3")
	m3 := newRPCServer(c2, 6997)
	c3.Join([]string{"127.0.0.1:5995", "127.0.0.1:5996"})

	expected = []*pb.Node{
		&pb.Node{
			Name:      "c1",
			Address:   "127.0.0.1:6995",
			HashBegin: tools.HashPoint(0, 3),
			HashEnd:   tools.HashPoint(1, 3),
		},
		&pb.Node{
			Name:      "c2",
			Address:   "127.0.0.1:6996",
			HashBegin: tools.HashPoint(1, 3),
			HashEnd:   tools.HashPoint(2, 3),
		},
		&pb.Node{
			Name:      "c3",
			Address:   "127.0.0.1:6997",
			HashBegin: tools.HashPoint(2, 3),
			HashEnd:   tools.HashPoint(3, 3),
		},
	}
	// All three nodes should agree about the member list
	assert.Equal(t, expected, c1.GetMembers())
	assert.Equal(t, expected, c2.GetMembers())
	assert.Equal(t, expected, c3.GetMembers())

	assert.Equal(t, 0, m1.Replications)
	assert.Equal(t, 0, m2.Replications)
	assert.Equal(t, 0, m3.Replications)
}

// mockRPCServer is a fake RPC server we use for this test.
type mockRPCServer struct {
	cluster      *Cluster
	Replications int
}

func (r *mockRPCServer) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	return r.cluster.AddNode(req), nil
}

func (r *mockRPCServer) Replicate(ctx context.Context, req *pb.ReplicateRequest) (*pb.ReplicateResponse, error) {
	r.Replications++
	return &pb.ReplicateResponse{Success: true}, nil
}

// newRPCServer creates a new mockRPCServer, starts a gRPC server running it, and returns it.
// It's not possible to stop it again...
func newRPCServer(cluster *Cluster, port int) *mockRPCServer {
	m := &mockRPCServer{cluster: cluster}
	s := grpc.NewServer()
	pb.RegisterRpcServerServer(s, m)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", port, err)
	}
	go s.Serve(lis)
	return m
}
