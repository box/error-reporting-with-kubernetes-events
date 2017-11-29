package main

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
)

func main() {
	var _ *kubernetes.Clientset
	fmt.Println("vim-go")
}
