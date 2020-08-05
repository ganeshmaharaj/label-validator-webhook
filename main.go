package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//  "k8s.io/apimachinery/pkg/runtime"
	//	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	whhttp "github.com/slok/kubewebhook/pkg/http"
	"github.com/slok/kubewebhook/pkg/log"
	whctx "github.com/slok/kubewebhook/pkg/webhook/context"
	validatingwh "github.com/slok/kubewebhook/pkg/webhook/validating"
)

var machines map[string]map[string]int
var owner_file = "/etc/node_owners/owner_file.yaml"

const nowner = "nodeowner"

func validateLabeling(ctx context.Context, obj metav1.Object) (bool, validatingwh.ValidatorResult, error) {
	var res validatingwh.ValidatorResult
	var reqobj corev1.Node

	// Validate the input requests
	// Make sure that the request is for a node in the cluster and that we are
	// able to view the final state of the object, a node in this case

	request := whctx.GetAdmissionRequest(ctx)
	if request == nil {
		fmt.Printf("Unable to read the request")
		res.Valid = false
		res.Message = fmt.Sprintf("Unable to read request")
		return false, res, nil
	}

	if request.Kind.Kind != "Node" {
		fmt.Printf("Not related to nodes. Skip")
		res.Valid = true
		res.Message = fmt.Sprintf("Not related to nodes")
		return true, res, nil
	}

	defer func() {
		reqobj = corev1.Node{}
	}()

	if ok := json.Unmarshal(request.Object.Raw, &reqobj); ok != nil {
		fmt.Printf("Unable to read the request: %s. Rejecting it!", ok)
		res.Valid = false
		res.Message = fmt.Sprintf("Unable to read the request: %s. Rejecting it!", ok)
		return false, res, ok
	}

	// Populate machine struct owned by the various teams
	// We construct the map structure after reading the owner_file.yaml which
	// allows us to deny request on machines not owned by that user.

	yamlfile, err := ioutil.ReadFile(owner_file)
	if err != nil {
		fmt.Printf("Unable to read the owners file %s: Error: %s", owner_file, err)
		res.Valid = false
		res.Message = fmt.Sprintf("Unable to read the owners file %s: Error: %s", owner_file, err)
		return false, res, err
	}

	if err = yaml.Unmarshal([]byte(yamlfile), &machines); err != nil {
		fmt.Printf("Unable to unmarshal the yaml file: %s Error: %s", owner_file, err)
		res.Valid = false
		res.Message = fmt.Sprintf("Unable to unmarshal the yaml file: %s Error: %s", owner_file, err)
		return false, res, err
	}

	//Check if the user is someone we care about. if not we allow all requests.
	// This change will allow users that are not tracked in the file (privileged
	// users) to continue with any operation on any host in the cluster.

	if _, ok := machines[request.UserInfo.Username]; !ok {
		fmt.Printf("Not a user we are concerned about. Allowing request")
		res.Valid = true
		res.Message = fmt.Sprintf("Not a user we are concerned about. Allowing request")
		return true, res, nil
	}

	//If the machine is not owned by the user attempting to label it, we will reject the operation. We do this using three kinds of checks.
	// 1. Make sure the mandatory label is not being deleted
	// 2. Make sure that the machine is owned by the user
	// 3. Make sure the current nodeowner is set to the user
	if _, ok := reqobj.ObjectMeta.Labels[nowner]; !ok {
		fmt.Printf("User %s is trying to remove mandatory label %s. Rejecting request", request.UserInfo.Username, nowner)
		res.Valid = false
		res.Message = fmt.Sprintf("User %s is trying to remove mandatory label %snode. Rejecting request", request.UserInfo.Username, nowner)
		return false, res, err
	}

	if _, ok := machines[request.UserInfo.Username][reqobj.ObjectMeta.Name]; ok {
		if reqobj.ObjectMeta.Labels[nowner] != request.UserInfo.Username {
			fmt.Printf("Machine %s is owned by user %s, but currently assigned to someone else. Rejecting request", reqobj.ObjectMeta.Name, request.UserInfo.Username)
			res.Valid = false
			res.Message = fmt.Sprintf("Machine %s is owned by user %s, but currently assigned to someone else. Rejecting request", reqobj.ObjectMeta.Name, request.UserInfo.Username)
			return false, res, err
		}
	} else {
		fmt.Printf("Machine %s is not owned by user %s. Rejecting request", reqobj.ObjectMeta.Name, request.UserInfo.Username)
		res.Valid = false
		res.Message = fmt.Sprintf("Machine %s is not owned by user %s. Rejecting request", reqobj.ObjectMeta.Name, request.UserInfo.Username)
		return false, res, err
	}

	// Iterate through the labels to make sure that the labels the user is trying
	// to work through has the username as prefix. For eg: all machines owned by
	// user 'xyz' have their labels prefixed with 'xyz'.
	//Exceptions: *kubernetes.io*, nodeowner

	for lbl := range reqobj.ObjectMeta.Labels {
		if strings.Contains(lbl, ".io") || strings.Contains(lbl, nowner) {
			continue
		}

		if !strings.HasPrefix(lbl, request.UserInfo.Username) {
			res.Valid = false
			res.Message = fmt.Sprintf("%s cannot apply labels that does not begin with %s", request.UserInfo.Username, request.UserInfo.Username)
			return false, res, nil
		}
	}

	res.Valid = true
	res.Message = fmt.Sprintf("Machine %s owned by user %s. Allow request", reqobj.ObjectMeta.Name, request.UserInfo.Username)
	return true, res, nil
}

type config struct {
	certFile string
	keyFile  string
}

func initFlags() *config {
	cfg := &config{}

	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fl.StringVar(&cfg.certFile, "tls-cert-file", "", "TLS certificate file")
	fl.StringVar(&cfg.keyFile, "tls-key-file", "", "TLS key file")

	if ok := fl.Parse(os.Args[1:]); ok != nil {
		fmt.Printf("Unable to parse user args")
	}
	return cfg
}

func main() {
	logger := &log.Std{Debug: true}

	cfg := initFlags()

	vt := validatingwh.ValidatorFunc(validateLabeling)

	vcfg := validatingwh.WebhookConfig{
		Name: "labelingvalidator",
		Obj:  &corev1.Node{},
	}
	valh, err := validatingwh.NewWebhook(vcfg, vt, nil, nil, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook: %s", err)
		os.Exit(1)
	}

	whHandler, err := whhttp.HandlerFor(valh)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook handler: %s", err)
		os.Exit(1)
	}
	logger.Infof("Listening on: 8080")
	err = http.ListenAndServeTLS(":8080", cfg.certFile, cfg.keyFile, whHandler)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error serving webhook: %s", err)
		os.Exit(1)
	}
}
