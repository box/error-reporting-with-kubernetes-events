package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
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

// // watchPkis uses the http watch endpoint to "react" to changes to the
// // pkis. A goroutine with an infinite loop is started that continously
// // waits on the http watch endpoint. The errors are logged, the received pki's
// // are insered into the channel that is returned.
// func (r *thirdPartyResourceReader) watchPkis() <-chan PkiEvent {

// httpEndpoint := r.config.Host + r.pkisWatchEndpoint

// events := make(chan PkiEvent)
// go func() {
// for {
// resp, err := r.httpClient.Get(httpEndpoint)
// if err != nil {
// handleHttpGetError(err, r.retrySeconds)
// continue
// } else if resp.StatusCode != 200 {
// handleHttpGetError(errors.New("Invalid status code: "+resp.Status),
// r.retrySeconds)
// continue
// }

// decoder := json.NewDecoder(resp.Body)

// // Incrementally parse the body as long as it grows, It looks like this
// // is the behavior in watch endpoints
// for {
// var event PkiEvent
// err = decoder.Decode(&event)
// if err != nil {
// handleJsonDecodeError(err)
// break
// }
// events <- event
// }
// resp.Body.Close()

// glog.Infof("Completed one Pki reception and decode")
// }
// }()
// return events
// }

func main() {

	kubeConfig, err := restclient.InClusterConfig()
	if err == nil {
		glog.Fatalf("Could not get kubeconfig", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	restClient := kubeClient.RESTClient()
	_ = restClient

	fmt.Println("vim-go")
}
