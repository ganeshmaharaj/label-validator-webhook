# Label Validator Admission Controller

A simple admission controller that allows only a certain labels can be applied to a certain pre-defined nodes(defined in a configmap) in the cluster.
This controller has been stolen with pride and repurposed from https://github.com/mcastelino/kata-webhook

## How to use the admission controller

```
docker build -t gmmaha/labeling-validator
./create_certs.sh
kubectl apply -f deploy/
```

## What gets installed
* Create secrets with the certificates for the admission controller to work
* Create a deployment with the admission controller
  * This pod depends on an existance of a configmap named `machine-owners-list`
  * The configmap should contain data file named `owner_file.yaml` with contents similar to what is shown below.
* Create a service that exposes port 8080 to 443 of the service IP.

## External Dependencies
A configmap that contains the users and list of machines that they own.

This is a good example of the config map
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: machine-owners-list
data:
  owner_file.yaml: |-
    team1:
      machine1:
      machine2:
    team2:
      machine3:
      machine4:
```

## What does it do
* Rejects labels on nodes that are not owned by the owner mentioned in the configmap
* Rejects users from labeling machines with labels that do not start with their username
* Rejects user from remove label `username+"node"` from the node they own. That is a label set by the administrator and is used with PodNodeSelector admission plugin 
