package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net"
	"strings"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	pb "github.com/akos011221/driftscape/proto"
)

var (
	rdb    *redis.Client
	domain = "default.svc.cluster.local"
)

func main() {
	// Connect to Redis for terrain storage
	// Example: "region:2,4" -> "plains with a hill"
	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("redis.%s:6379", domain),
	})
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		fmt.Println("Redis connection failed:", err)
	}

	// Start gRPC server on :8081
	// Listens for Coordinator calls to region services
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		fmt.Println("Failed to listen:", err)
		return
	}
	s := grpc.NewServer()
	pb.RegisterRegionServiceServer(s, &regionServer{})
	fmt.Println("Region running on :8081")
	if err := s.Serve(lis); err != nil {
		fmt.Println("Failed to serve:", err)
	}
}

type regionServer struct {
	pb.UnimplementedRegionServiceServer
}

func (s *regionServer) GetDescription(ctx context.Context, pos *pb.Position) (*pb.Description, error) {
	// Generate terrain based on x,y
	// Example: (2,4) -> "plains with a hill"
	x, y := int(pos.X), int(pos.Y)
	terrain := generateTerrain(x, y)

	// Save to Redis
	key := fmt.Sprintf("region:%d,%d", x, y)
	rdb.Set(context.Background(), key, terrain, 0)

	return &pb.Description{Terrain: terrain}, nil
}

func generateTerrain(x, y int) string {
	// Seed randomness with x,y for consistency
	// Example: (2,4) always gets same base terrain
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%d,%d", x, y)))
	seed := h.Sum32()
	r := newRand(int64(seed))

	// Base terrain types
	bases := []string{"forest", "plains", "hills", "swamp", "coast", "river"}
	base := bases[r.Intn(len(bases))]

	// Add features with border sync
	// Example: If (2,3) has a river south, (2,4) reflects it
	feature := ""
	if r.Float32() < 0.3 { // 30% chance of a feature
		features := []string{"", "with a cave", "with ancient ruins", "with a waterfall"}
		feature = " " + features[r.Intn(len(features))]
		if strings.Contains(feature, "river") {
			// Check south neighbor for river
			southKey := fmt.Sprintf("region:%d,%d", x, y-1)
			southTerrain, _ := rdb.Get(context.Background(), southKey).Result()
			if strings.Contains(southTerrain, "river") {
				feature = " with a river flowing south"
			}
		}
	}
	// Combine for richer description
	// Example: "plains with a river flowing south"
	return base + feature
}

// newRand creates a seeded random generator
func newRand(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}
