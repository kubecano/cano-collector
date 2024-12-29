package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	log.Println("Starting cano-server...")

	// Initialize Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error creating in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes clientset: %v", err)
	}

	// Create a context that cancels on SIGINT/SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Println("Shutting down cano-server...")
		cancel()
	}()

	// Start periodic logging
	logTicker := time.NewTicker(1 * time.Minute)
	defer logTicker.Stop()

	for {
		select {
		case <-logTicker.C:
			logMessage := fmt.Sprintf("cano-server is running at %v", time.Now().Format(time.RFC3339))
			log.Println(logMessage)

			// Example: Interact with Kubernetes API
			_, err := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
			if err != nil {
				log.Printf("Error listing pods: %v", err)
			} else {
				log.Println("Successfully interacted with Kubernetes API.")
			}

		case <-ctx.Done():
			log.Println("Exiting cano-server...")
			return
		}
	}
}
