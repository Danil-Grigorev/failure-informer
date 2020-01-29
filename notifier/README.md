# Sample kubebuilder application - Event Notifier

Sample kubebuilder cloud-native app, which sends email notifications on pod failure (sort of).

This is done by using existing tools provided by `kubebuilder` framework, via extending existing Kubernetes [object](https://kubernetes.io/docs/concepts/#kubernetes-objects) behavior - [`v1.Event`](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.10/#event-v1-core), and creating our own Custom Resource [`(CR)`](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) - named `Notifier`

# Prerequisites

1. Kubernetes 1.17+ / Openshift 3.11+
2. Golang 1.13
3. Kubebuilder 2.0.0
4. Kustomize

# Install

1. Deploy a cluster
- [oc cluster up](https://docs.okd.io/latest/getting_started/administrators.html#installation-methods) (openshift)
- [Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/)
2. [Install](https://golang.org/doc/install) golang
3. [Install](https://book.kubebuilder.io/quick-start.html#installation) kubebuilder

# Cloud-native application behavior

![Expected behavior](./images/plan.png)

# Preparation

```bash
# Create initial project structure
kubebuilder init --domain email.notify.io --license apache2
# Scaffold our v1.Notifier CR
# Responding yes on both controller and resource creation
kubebuilder create api --group email --version v1 --kind Notifier

# Extending existing v1.Event kubernetes resource with our controller
# From provided options select only the controller scaffolding, as the resource will be already present
kubebuilder create api --group core --version v1 --kind Event
```
----------------------
There are several places, where the code for this workshop will be edited. For better orientation here is an overview:
1. [notifier_types](./api/v1/notifier_types.go) - Location of the `v1.Notifier` CR structures, helper functions and filters.
2. [notifier_controller](./controllers/notifier_controller.go) - Specifically `Reconcile` function, is the place where the controller logic is located at. This part will be executed every time any of: `Create`|`Update`|`Delete`|`Generic` events are captured by the controller, related to our `v1.Notifier` `CR`.
3. [event_controller](./controllers/event_controller.go) - Our extension for `v1.Event` behavior, with another controller. Notice usage of predicates, to filter incoming events [here](https://github.com/Danil-Grigorev/failure-informer/blob/76eaf33ddc7849f49259830b1def8134468221c9/notifier/controllers/event_controller.go#L85)
4. [event_predicate](./controllers/event_predicate.go) - This file is specifically dedicated to filtering incoming events for `v1.Event` resource, which should trigger our custom [event_controller](./controllers/event_controller.go) reconciliation run.

## Example CR - `email.notify.io/v1.Notifier`

```yaml
apiVersion: email.notify.io/v1
kind: Notifier
metadata:
  name: notifier-sample
  namespace: test
spec:
  # Add fields here
  email: test@test.com
  filters:
  - BackOff
```

# Executing the controller's code

## Locally
```bash
make
make manifests
make install
make run
```
- Create a namespace, faulty pod and a `Notifier` CR from [samples](./samples) by running
```bash
oc create -f ./samples
```

## Production

- First you will need to edit container image url, where the final application will be published, then later on pulled and used in production - [here](https://github.com/Danil-Grigorev/failure-informer/blob/76eaf33ddc7849f49259830b1def8134468221c9/notifier/config/default/manager_image_patch.yaml#L10)
- Same thing for the `Makefile` - [here](https://github.com/Danil-Grigorev/failure-informer/blob/76eaf33ddc7849f49259830b1def8134468221c9/notifier/Makefile#L3)
- Then:
```bash
make
make manifests
make install
make docker-build
make docker-push
make deploy
```
- Create a namespace, faulty pod and a `Notifier` CR from [samples](./samples) by running
```bash
oc create -f ./samples
```

# Sources
- [kubebuilder book](https://book.kubebuilder.io/)
- [Programming kubernetes](https://learning.oreilly.com/library/view/programming-kubernetes/9781492047094/)
