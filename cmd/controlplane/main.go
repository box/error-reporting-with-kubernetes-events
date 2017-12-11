package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"os"

	pkiV1 "github.com/box/error-reporting-with-kubernetes-events/pkg/apis/box.com/v1"
	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// isInSlice returns true if the string e exists in slice s
func isInSlice(s []string, e string) bool {
	for _, x := range s {
		if x == e {
			return true
		}
	}
	return false
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

// doPkiProcessing is a placeholder function for the business logic in this
// control plane application. Once a valid Pki object is found, additional
// processing would be triggered from this function. At the end the output is
// written back to the api server in a Secret so that the interested parties
// can mount it.
func doPkiProcessing(pki pkiV1.Pki, kubeClient *kubernetes.Clientset) {
	//
	// .....
	// Omitted extra processing for brevity
	//

	// Generate a Secret with the processing result and write back to the
	// api server. The application does a secret mount and retrieves this
	// data.

	// A placeholder empty secret. In real life, the data in this
	// secret would contain the generated pki information.
	emptySecretSpec := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pki.Name,
			Namespace: pki.Namespace,
		},
	}
	_, err := kubeClient.CoreV1().Secrets(pki.Namespace).Create(
		&emptySecretSpec)
	if err != nil {
		glog.Fatalf("Could not generate Secret: %v", err)
	}
}

// handlePkiEvents waits for new Pki's to be added or existing ones to be modified.
// After it such event, it does parameter checking and some external processing.
func handlePkiEvents(pkiEvents <-chan pkiV1.PkiEvent, allowedNames []string,
	kubeClient *kubernetes.Clientset) {

	for e := range pkiEvents {
		glog.Infof("Seen PkiEvent : %v", e)
		t := e.Type
		pki := e.Object
		if t == watch.Added || t == watch.Modified {
			serviceName := pki.Spec.ServiceName
			if isInSlice(allowedNames, serviceName) {
				// The input parameters are ok. Do extra processing ...
				doPkiProcessing(pki, kubeClient)
			} else {
				// Found some unexpected parameters in the pki. Generate an error
				// message using kubernetes events.
			}

		} else {
			// There are other types of changes, Deleted, Modified etc that we do
			// not care in this example.
			glog.Infof("Received an unhandled pki change: %s for %v. Ignoring!", t, pki)
		}
		glog.Flush()
	}
}

func main() {

	flag.Parse()

	allowedNames := allowedNames()

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
	handlePkiEvents(pkiEvents, allowedNames, kubeClient)
}
