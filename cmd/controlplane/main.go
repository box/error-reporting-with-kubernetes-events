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
const pkisWatchPath string = "/apis/box.com/v1/pkis"

// allowedNames returns the allowed serviceName values in PKI objects.
// The file in allowedNamesFilePath is read and its contents are returned in a
// slice. One line each slice entry.
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

// watchPkis uses the HTTP watch endpoint to "react" to changes to the
// PKIs. A goroutine with an infinite loop is started that continously
// waits on the HTTP watch endpoint. The errors are logged, the received PKIs
// are inserted into the channel that is returned.
func watchPkis(restClient restclient.Interface) <-chan pkiV1.PkiChange {

	events := make(chan pkiV1.PkiChange)
	go func() {
		for {
			// TODO: Instead of this direct HTTP request, the workqueue and
			// informer pattern could be used. Similar to here
			// https://github.com/kubernetes/sample-controller/blob/258eead08702028194ade1c0a7f958c837d6f081/controller.go#L122
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
				var event pkiV1.PkiChange
				err = decoder.Decode(&event)
				if err != nil {
					glog.Errorf("JSON decode error %v.", err)
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
// control plane application. Once a valid PKI object is found, additional
// processing would be triggered from this function. At the end the output is
// written back to the API server in a Secret so that the interested parties
// can mount it.
func doPkiProcessing(pki pkiV1.Pki, kubeClient *kubernetes.Clientset) {
	//
	// .....
	// Omitted extra processing for brevity
	//

	// Generate a Secret with the processing result and write back to the
	// API server. The application does a Secret mount and retrieves this
	// data.

	// A placeholder empty Secret. In real life, the data in this
	// Secret would contain the generated PKI information.
	emptySecretSpec := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pki.Name,
			Namespace: pki.Namespace,
		},
	}

	// Replace the PKI
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

// postEventAboutPki uses Kubernetes Events to write an error message to the event
// steam of pods in the same namespace as the PKI.
func postEventAboutPki(pki pkiV1.Pki, kubeClient *kubernetes.Clientset,
	recorder record.EventRecorder, allowedNames []string) {

	pods, err := kubeClient.CoreV1().Pods(pki.Namespace).List(metav1.ListOptions{})
	if err != nil {
		glog.Fatalf(" Could not list pods in namespace %s: %v", pki.Namespace, err)
	}
	for _, pod := range pods.Items {
		// TODO: Even in the same namespace, there may be some pods that do not use
		// this PKI. Ideally we only want to send this message to the relevant pods.
		// One needs to inspect the volumeMounts done by the containers in pods.
		// If a container does a volumeMount with the same name, only then post this
		// error message to that Pod's lifecycle. This is left as an exercise for the
		// reader.

		ref, err := reference.GetReference(scheme.Scheme, &pod)
		if err != nil {
			glog.Fatalf("Could not get reference for pod %v: %v\n",
				pod.Name, err)
		}
		recorder.Event(ref, v1.EventTypeWarning, "PKI ServiceName error",
			fmt.Sprintf("ServiceName: %s in PKI: %s is not found in"+
				" allowedNames: %s", pki.Spec.ServiceName, pki.Name,
				allowedNames))
	}
}

// handlePkiChanges waits for new PKIs to be added or existing ones to be modified.
// When a change happens, it does parameter checking and some external processing.
func handlePkiChanges(pkiEvents <-chan pkiV1.PkiChange, allowedNames []string,
	kubeClient *kubernetes.Clientset, recorder record.EventRecorder) {

	for e := range pkiEvents {
		glog.Infof("Seen PkiChange : %v", e)
		t := e.Type
		pki := e.Object
		if t == watch.Added || t == watch.Modified {
			serviceName := pki.Spec.ServiceName
			if isInSlice(allowedNames, serviceName) {
				// The input parameters are ok. Do extra processing ...
				doPkiProcessing(pki, kubeClient)
			} else {
				// Found some unexpected parameters in the PKI. Generate an error
				// message using kubernetes events.
				postEventAboutPki(pki, kubeClient, recorder, allowedNames)
			}
		} else {
			// There are other types of changes, Deleted, etc that we do
			// not care in this example.
			glog.Infof("Received an unhandled PKI change: %s for %v. Ignoring!", t, pki)
		}
		glog.Flush()
	}
}

// eventRecorder returns an EventRecorder type that can be
// used to post Events to different object's lifecycles.
func eventRecorder(
	kubeClient *kubernetes.Clientset) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{
			Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(
		scheme.Scheme,
		v1.EventSource{Component: "controlplane"})
	return recorder
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
	recorder := eventRecorder(kubeClient)
	pkiEvents := watchPkis(restClient)
	handlePkiChanges(pkiEvents, allowedNames, kubeClient, recorder)
}
