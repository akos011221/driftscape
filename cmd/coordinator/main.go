package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	rdb       *redis.Client
	clientset *kubernetes.Clientset
	domain    = "default.svc.cluster.local"
)

func main() {
	// Connect to Redis (assumes redis.<domain>:6379 in Kubernetes)
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

	fmt.Println("Coordinator running on :8080")
	http.ListenAndServe(":8080", nil)
}

// lookHandler responds when the Client asks what's at (x,y)
func lookHandler(w http.ResponseWriter, r *http.Request) {
	// Gets x and y from the request (e.g., "?x=0&y=1")
	x, y, err := getXY(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// If new or re-visited region, spawn it
	key := fmt.Sprintf("region:%d,%d", x, y)
	regionData, err := rdb.Get(context.Background(), key).Result()
	if err == redis.Nil || !regionExists(x, y) {
		regionData = spawnRegion(x, y)
	} else if err != nil {
		http.Error(w, "Redis error", 500)
		return
	}

	// Ask the region pod for its description
	podName := fmt.Sprintf("region-%d-%d", x, y)
	url := fmt.Sprintf("http://%s.%s:8081/desc", podName, domain)
	resp, err := http.Get(url)
	if err != nil {
		// Fallback description if the pod isn't ready yet
		// Prints a simpler message using the data from Redis
		fmt.Fprintf(w, "You're in a %s at (%d,%d)", regionData, x, y)
		return
	}
	defer resp.Body.Close()
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	w.Write(buf[:n])
}

// moveHandler handles when you move to a new spot
func moveHandler(w http.ResponseWriter, r *http.Request) {
	x, y, err := getXY(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Get old position from Redis
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

	// If new or re-visited region, spawn it
	key := fmt.Sprintf("region:%d,%d", x, y)
	regionData, err := rdb.Get(context.Background(), key).Result()
	if err == redis.Nil || !regionExists(x, y) {
		regionData = spawnRegion(x, y)
	} else if err != nil {
		http.Error(w, "Redis error", 500)
		return
	}

	// Update position in Redis
	rdb.Set(context.Background(), "user:position", fmt.Sprintf("%d,%d", x, y), 0)

	// Get description from new region
	podName := fmt.Sprintf("region-%d-%d", x, y)
	url := fmt.Sprintf("http://%s.%s:8081/desc", podName, domain)
	resp, err := http.Get(url)
	if err != nil {
		// Fallback description if the pod isn't ready yet
		// Prints a simpler message using data from Redis
		fmt.Fprintf(w, "You moved to a %s at (%d,%d)", regionData, x, y)
		return
	}
	defer resp.Body.Close()
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	w.Write(buf[:n])
}

// getXY extracts x,y from the HTTP request
func getXY(r *http.Request) (int, int, error) {
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

// spawnRegion creates a new region pod and saves it to Redis
func spawnRegion(x, y int) string {
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

// deleteRegion removes a region pod
func deleteRegion(x, y int) {
	podName := fmt.Sprintf("region-%d-%d", x, y)
	clientset.AppsV1().Deployments("default").Delete(context.Background(), podName, metav1.DeleteOptions{})
	clientset.CoreV1().Services("default").Delete(context.Background(), podName, metav1.DeleteOptions{})
}

// parsePosition splits "x,y" into numbers
func parsePosition(pos string) (int, int) {
	parts := strings.Split(pos, ",")
	x, _ := strconv.Atoi(parts[0])
	y, _ := strconv.Atoi(parts[1])
	return x, y
}

func regionExists(x, y int) bool {
	podName := fmt.Sprintf("region-%d-%d", x, y)
	_, err := clientset.AppsV1().Deployments("default").Get(context.Background(), podName, metav1.GetOptions{})
	return err == nil
}

// int32Ptr creates a pointer to an int32 value
func int32Ptr(i int32) *int32 { return &i }
