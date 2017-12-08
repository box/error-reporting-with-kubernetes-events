package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
)

const allowedNamesFilePath string = "/AllowedNames"

// allowedNames returns the allowed serviceName values in PKi objects.
// The file in allowedNamesFilePath is read and its contents are retuned in a
// slide. One line each slice entry.
func allowedNames() []string {
	file, err := os.Open(allowedNamesFilePath)
	if err != nil {
		glog.Fatalf("Could not open file: %s", allowedNamesFilePath)
	}
	defer file.Close()

	rv := []string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) != 0 {
			rv = append(rv, line)
		}
	}

	if err := scanner.Err(); err != nil {
		glog.Fatal(err)
	}

	return rv
}

func main() {
	var _ *kubernetes.Clientset
	fmt.Println("vim-go")
}
