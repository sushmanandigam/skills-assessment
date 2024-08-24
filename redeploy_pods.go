package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Load kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Fatalf("Error loading kubeconfig: %v", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}

	// Get all pods
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing pods: %v", err)
	}

	// Iterate over pods and redeploy those with "database" in the name
	for _, pod := range pods.Items {
		if containsDatabase(pod.Name) {
			fmt.Printf("Redeploying pod: %s\n", pod.Name)

			// Get the deployment associated with the pod
			deployment, err := clientset.AppsV1().Deployments(pod.Namespace).Get(context.TODO(), pod.Labels["app"], metav1.GetOptions{})
			if err != nil {
				log.Fatalf("Error getting deployment: %v", err)
			}

			// Add an annotation to trigger a restart
			if deployment.Spec.Template.Annotations == nil {
				deployment.Spec.Template.Annotations = make(map[string]string)
			}
			deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = metav1.Now().String()

			// Update the deployment to trigger a restart
			_, err = clientset.AppsV1().Deployments(pod.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			if err != nil {
				log.Fatalf("Error updating deployment: %v", err)
			}
			fmt.Printf("Successfully restarted pod: %s\n", pod.Name)
		}
	}
}

// containsDatabase checks if the pod name contains the substring "database"
func containsDatabase(name string) bool {
	return strings.Contains(name, "database")
}
