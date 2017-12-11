package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"os"

	pkiV1 "github.com/box/error-reporting-with-kubernetes-events/pkg/apis/box.com/v1"
	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/watch"
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

// watchPkis uses the http watch endpoint to "react" to changes to the
// pkis. A goroutine with an infinite loop is started that continously
// waits on the http watch endpoint. The errors are logged, the received pki's
// are insered into the channel that is returned.
func watchPkis(restClient restclient.Interface) <-chan pkiV1.PkiEvent {

	pkisWatchPath := "/apis/box.com/v1/pkis"

	events := make(chan pkiV1.PkiEvent)
	go func() {
		for {
			request := restClient.Get().AbsPath(pkisWatchPath).Param("watch", "true")
			glog.Infof("HTTP Get request for: %s", request.URL())
			body, err := request.Stream()
			if err != nil {
				glog.Errorf("restClient Stream Error  %v.", err)
				continue
			}
			defer body.Close()

			decoder := json.NewDecoder(body)

			// Incrementally parse the body as long as it grows, It looks like this
			// is the behavior in watch endpoints
			for {
				var event pkiV1.PkiEvent
				err = decoder.Decode(&event)
				if err != nil {
					glog.Errorf("Json decode error %v.", err)
					break
				}
				events <- event
			}

			glog.Infof("Completed one Pki reception and decode")
		}
	}()
	return events
}

func main() {

	flag.Parse()
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		glog.Fatalf("Could not get kubeconfig: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %v", err)
	}

	restClient := kubeClient.RESTClient()

	pkiEvents := watchPkis(restClient)

	// Handle pki additions.
	for e := range pkiEvents {
		glog.Infof("Seen PkiEvent : %v", e)
		t := e.Type
		pki := e.Object
		if t == watch.Added || t == watch.Modified {
			// runStatus := NewRunStatus()
			// genStatus, _ := psg.Generate(&pki)
			// runStatus.Pki = pki
			// runStatus.GenerateStatus = genStatus
			// runStatus.RecordDuration()
			// httpHandlerData.prepend(runStatus)

		} else {
			// There are other types of changes, Deleted, Modified etc that we do
			// not care in this example.
			glog.Infof("Received an unhandled pki change: %s for %v. Ignoring!", t, pki)
		}
		glog.Flush()
	}
}
