package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	pkiV1 "github.com/box/error-reporting-with-kubernetes-events/pkg/apis/box.com/v1"
	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
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
			// TODO: Instead of this direct HTTP request, the workqueue and
			// informer pattern could be used. Similar to here
			// https://github.com/kubernetes/sample-controller/blob/master/controller.go#L122
			request := restClient.Get().AbsPath(pkisWatchPath).Param("watch", "true")
			glog.Infof("HTTP Get request for: %s", request.URL())
			body, err := request.Stream()
			if err != nil {
				glog.Errorf("restClient Stream Error  %v.", err)
				continue
			}
			defer body.Close()

			decoder := json.NewDecoder(body)

			// Incrementally parse the body as long as it grows.
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

	// Replace the Pki
	err := kubeClient.CoreV1().Secrets(pki.Namespace).Delete(pki.Name,
		&metav1.DeleteOptions{})
	if err != nil {
		glog.Infof("Could not delete Secret with name %s: %v\n", pki.Name, err)
	}

	_, err = kubeClient.CoreV1().Secrets(pki.Namespace).Create(
		&emptySecretSpec)
	if err != nil {
		glog.Fatalf("Could not generate Secret: %v", err)
	}
}

func postEventAboutPki(pki pkiV1.Pki, kubeClient *kubernetes.Clientset,
	recorder record.EventRecorder, allowedNames []string) error {

	pods, err := kubeClient.CoreV1().Pods(pki.Namespace).List(metav1.ListOptions{})
	if err != nil {
		glog.Fatalf(" Could not list pods in namespace %s: %v", pki.Namespace, err)
	}
	for _, pod := range pods.Items {
		// TODO: Even in the same namespace, there may be some pods that do not use
		// this pki. Ideally we only want to send this message to the relevant pods.
		// One needs to inspect the volumeMounts done by the containers in pods.
		// If a container does a volumeMount with the same name, only then post this
		// error message to that Pod's lifecycle. This is left as an excercise for the
		// reader.

		ref, err := reference.GetReference(scheme.Scheme, &pod)
		if err != nil {
			glog.Fatalf("Could not get refecence for pod %v: %v\n", pod.Name, err)
		}
		recorder.Event(ref, v1.EventTypeWarning, "pki ServiceName error",
			fmt.Sprintf("ServiceName: %s in pki: %s is not found in allowedNames: %s",
				pki.Spec.ServiceName, pki.Name, allowedNames))
	}
	return nil
}

// handlePkiEvents waits for new Pki's to be added or existing ones to be modified.
// After it such event, it does parameter checking and some external processing.
func handlePkiEvents(pkiEvents <-chan pkiV1.PkiEvent, allowedNames []string,
	kubeClient *kubernetes.Clientset, recorder record.EventRecorder) {

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
				postEventAboutPki(pki, kubeClient, recorder, allowedNames)
			}
		} else {
			// There are other types of changes, Deleted, Modified etc that we do
			// not care in this example.
			glog.Infof("Received an unhandled pki change: %s for %v. Ignoring!", t, pki)
		}
		glog.Flush()
	}
}

// eventRecorder returns an EventRecorder type that can be
// used to post Events to different object's lifecycles.
func eventRecorder(
	kubeClient *kubernetes.Clientset) (record.EventRecorder, error) {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{
			Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(
		scheme.Scheme,
		v1.EventSource{Component: "controlplane"})
	return recorder, nil
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
	recorder, err := eventRecorder(kubeClient)
	if err != nil {
		glog.Fatalf("Error getting EventRecorder: %v", err)
	}
	pkiEvents := watchPkis(restClient)
	handlePkiEvents(pkiEvents, allowedNames, kubeClient, recorder)
}
