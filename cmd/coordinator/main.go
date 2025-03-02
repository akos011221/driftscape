package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	pb "github.com/akos011221/driftscape/proto"
)

var (
	rdb       *redis.Client
	clientset *kubernetes.Clientset
	domain    = "default.svc.cluster.local"
)

func main() {
	// Connect to Redis for persistent storage
	// Example: redis.default.svc.cluster.local:6379 holds "user:position" -> "2,3"
	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("redis.%s:6379", domain), // Service DNS in K8s
	})
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic("Redis connection failed: " + err.Error())
	}

	// Connect to Kubernetes (in-cluster config)
	config, err := rest.InClusterConfig()
	if err != nil {
		panic("Kubernetes config failed: " + err.Error())
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic("Kubernetes client failed: " + err.Error())
	}

	http.HandleFunc("/look", lookHandler)
	http.HandleFunc("/move", moveHandler)
	http.HandleFunc("/position", positionHandler)

	fmt.Println("Coordinator running on :8080")
	http.ListenAndServe(":8080", nil)
}

func positionHandler(w http.ResponseWriter, r *http.Request) {
	// Send last known position to Client
	// Example: Client gets "2,3" from Redis
	pos, err := rdb.Get(context.Background(), "user:position").Result()
	if err == redis.Nil {
		fmt.Fprintf(w, "0,0") // Center, if no position
	} else if err != nil {
		http.Error(w, "Redis error", 500)
		return
	} else {
		fmt.Fprintf(w, pos) // Send last known position
	}
}

func lookHandler(w http.ResponseWriter, r *http.Request) {
	// Get x,y from Client request
	// Example: "?x=2&y=3" from "look" command
	x, y, err := getXY(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Check or spawn region in Redis/K8s
	// Example: "region:2,3" -> "forest" or spawn pod
	key := fmt.Sprintf("region:%d,%d", x, y)
	regionData, err := rdb.Get(context.Background(), key).Result()
	if err == redis.Nil || !regionExists(x, y) {
		regionData = spawnRegion(x, y)
	} else if err != nil {
		http.Error(w, "Redis error", 500)
		return
	}

	// Call Region pod via gRPC
	// Example: Dial "region-2-3:8081", get "forest with a river"
	podName := fmt.Sprintf("region-%d-%d", x, y)
	desc, err := getRegionDescription(podName, x, y)
	if err != nil {
		fmt.Fprintf(w, "You are in a %s at (%d,%d)", regionData, x, y)
		return
	}
	fmt.Fprintf(w, "You're in a %s at (%d,%d)", desc, x, y)
}

func moveHandler(w http.ResponseWriter, r *http.Request) {
	// Parse new position from Client
	// Example: "?x=2&y=4" from "move north"
	x, y, err := getXY(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Clean up old position’s pod
	// Example: Was at "2,3", now "2,4"—delete region-2-3
	oldPos, err := rdb.Get(context.Background(), "user:position").Result()
	if err != nil && err != redis.Nil {
		http.Error(w, "Redis error", 500)
		return
	}

	// If there's an old position, clean up its pod
	if oldPos != "" && oldPos != fmt.Sprintf("%d,%d", x, y) {
		oldX, oldY := parsePosition(oldPos)
		deleteRegion(oldX, oldY)
	}

	// Check or spawn new region
	// Example: "region:2,4" -> "plains" or spawn pod
	key := fmt.Sprintf("region:%d,%d", x, y)
	regionData, err := rdb.Get(context.Background(), key).Result()
	if err == redis.Nil || !regionExists(x, y) {
		regionData = spawnRegion(x, y)
	} else if err != nil {
		http.Error(w, "Redis error", 500)
		return
	}

	// Save new position
	// Example: "user:position" -> "2,4" in Redis
	rdb.Set(context.Background(), "user:position", fmt.Sprintf("%d,%d", x, y), 0)

	// Get description via gRPC
	// Example: "region-2-4:8081" -> "plains with a hill"
	podName := fmt.Sprintf("region-%d-%d", x, y)
	desc, err := getRegionDescription(podName, x, y)
	if err != nil {
		fmt.Fprintf(w, "You moved to a %s at (%d,%d)", regionData, x, y)
		return
	}
	fmt.Fprintf(w, "You moved to a %s at (%d,%d)", desc, x, y)
}

func getXY(r *http.Request) (int, int, error) {
	// Parse x,y from query params
	// Example: "?x=2&y=4" -> x=2, y=4
	xStr := r.URL.Query().Get("x")
	yStr := r.URL.Query().Get("y")
	x, err := strconv.Atoi(xStr)
	if err != nil {
		return 0, 0, fmt.Errorf("Bad x!")
	}
	y, err := strconv.Atoi(yStr)
	if err != nil {
		return 0, 0, fmt.Errorf("Bad y!")
	}
	return x, y, nil
}

func spawnRegion(x, y int) string {
	// Create a new region pod with HPA
	// Example: Spawns "region-2-4" pod + service in OKE
	podName := fmt.Sprintf("region-%d-%d", x, y)
	// Sanitize x, y for labels (replace negative with 'n')
	xLabel := strconv.Itoa(x)
	if x < 0 {
		xLabel = "n" + strconv.Itoa(-x)
	}
	yLabel := strconv.Itoa(y)
	if y < 0 {
		yLabel = "n" + strconv.Itoa(-y)
	}

	// Check if Deployment exists, delete if broken
	_, err := clientset.AppsV1().Deployments("default").Get(context.Background(), podName, metav1.GetOptions{})
	if err == nil {
		deleteRegion(x, y) // Clean up stale Deployment
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "region",
					"x":   xLabel,
					"y":   yLabel,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "region",
						"x":   xLabel,
						"y":   yLabel,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "region",
							Image: "orbanakos2312/driftscape-region",
							Env: []corev1.EnvVar{
								{Name: "REGION_X", Value: strconv.Itoa(x)},
								{Name: "REGION_Y", Value: strconv.Itoa(y)},
							},
							Ports: []corev1.ContainerPort{{ContainerPort: 8081}},
							//ReadinessProbe: &corev1.Probe{
							//	ProbeHandler: corev1.ProbeHandler{
							//		GRPC: &corev1.GRPCAction{
							//			Port: 8081,
							//		},
							//	},
							//	InitialDelaySeconds: 2, // Wait 2s before first check
							//	PeriodSeconds:       2, // Check every 2s
							//	FailureThreshold:    3, // Fail after 3 tries
							//},
							Resources: corev1.ResourceRequirements{ // For HPA
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resourceMustParse("100m"),
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment in default namespace
	_, err = clientset.AppsV1().Deployments("default").Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Failed to spawn region:", err)
	}

	// Add HPA for scaling
	// Example: Scales region-2-4 if CPU hits 50%
	hpa := &autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName + "-hpa",
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       podName,
				APIVersion: "apps/v1",
			},
			MinReplicas:                    int32Ptr(1),
			MaxReplicas:                    3, // Max 3 pods per region
			TargetCPUUtilizationPercentage: int32Ptr(50),
		},
	}
	_, err = clientset.AutoscalingV1().HorizontalPodAutoscalers("default").Create(context.Background(), hpa, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Failed to create HPA:", err)
	}

	// Create Service for the pod
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "region",
				"x":   xLabel,
				"y":   yLabel,
			},
			Ports: []corev1.ServicePort{
				{Port: 8081, TargetPort: intstr.FromInt(8081)},
			},
		},
	}
	_, err = clientset.CoreV1().Services("default").Create(context.Background(), service, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Failed to create service:", err)
	}

	// Save basic region type to Redis (pod will refine it)
	regionType := "unknown" // Placeholder, pod sets real type
	rdb.Set(context.Background(), fmt.Sprintf("region:%d,%d", x, y), regionType, 0)
	return regionType
}

func deleteRegion(x, y int) {
	// Remove a region pod and its HPA
	// Example: Deletes "region-2-3" and "region-2-3-hpa"
	podName := fmt.Sprintf("region-%d-%d", x, y)
	clientset.AppsV1().Deployments("default").Delete(context.Background(), podName, metav1.DeleteOptions{})
	clientset.CoreV1().Services("default").Delete(context.Background(), podName, metav1.DeleteOptions{})
	clientset.AutoscalingV1().HorizontalPodAutoscalers("default").Delete(context.Background(), podName+"-hpa", metav1.DeleteOptions{})
}

func parsePosition(pos string) (int, int) {
	// Split "x,y" into numbers
	// Example: "2,3" -> x=2, y=3
	parts := strings.Split(pos, ",")
	x, _ := strconv.Atoi(parts[0])
	y, _ := strconv.Atoi(parts[1])
	return x, y
}

func regionExists(x, y int) bool {
	// Check if a region pod exists
	// Example: Looks for "region-2-4" in OKE
	podName := fmt.Sprintf("region-%d-%d", x, y)
	_, err := clientset.AppsV1().Deployments("default").Get(context.Background(), podName, metav1.GetOptions{})
	return err == nil
}

func int32Ptr(i int32) *int32 { return &i }

func resourceMustParse(s string) resource.Quantity {
	// Parse resource strings for HPA
	// Example: "100m" -> 0.1 CPU (100 milliCPU)
	q, _ := resource.ParseQuantity(s)
	return q
}

func getRegionDescription(podName string, x, y int) (string, error) {
	// Connect to Region pod via gRPC
	// Example: Dials "region-2-4:8081", sends x=2, y=4
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s.%s:8081", podName, domain),
		grpc.WithTransportCredentials(insecure.NewCredentials()), // No TLS for simplicity
	)
	if err != nil {
		return "", fmt.Errorf("failed to connect to %s: %v", podName, err)
	}
	defer conn.Close()

	client := pb.NewRegionServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call GetDescription
	// Example: Gets "forest with a river" for (2,4)
	resp, err := client.GetDescription(ctx, &pb.Position{X: int32(x), Y: int32(y)})
	if err != nil {
		return "", fmt.Errorf("failed to get description from %s: %v", podName, err)
	}
	return resp.Terrain, nil
}
