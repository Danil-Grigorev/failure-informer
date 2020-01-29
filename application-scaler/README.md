# Sample operator build with kubebuilder

This repo contains a sample operator build by using toolkit provided by the [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) framework. The idea is to create a simplified version of Kubenetes [`Deployment`](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/), and call it `AppScaler`.

# Prerequisites

1. Kubernetes 1.17+ / Openshift 3.11+ cluster
2. Unpack and install `oc` binary - [here](https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz)
2. Golang 1.13
3. Kubebuilder 2.0.0

# Install

1. Deploy a cluster
- [oc cluster up](https://docs.okd.io/latest/getting_started/administrators.html#installation-methods) (openshift)
- [Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/)
2. [Install](https://golang.org/doc/install) golang
3. [Install](https://book.kubebuilder.io/quick-start.html#installation) kubebuilder

# Steps to create a CR

We need to scaffold our operator code with kubebulder. Recommend to have a look at kubebuilder [Getting Started](https://github.com/kubernetes-sigs/kubebuilder#getting-started) to get a grasp.

```bash
kubebuilder init --domain example.com --license apache2
kubebuilder create api --group deploy --version v1 --kind AppScaler
```

This will generate our initial project structure, similar to one hosted in this repo. Notice initial template for our CR in the `./config/samples` directory. You may edit it at any moment to match the CR specification in the [appscaler_types](./api/v1beta1/appscaler_types.go), and then deploy it in your cluster.

# CR - `example.com/v1.AppScaler`

```yaml
apiVersion: sample.example.com/v1
kind: AppScaler
metadata:
  name: appscaler-sample
  namespace: test
spec:
  replicas: 2
  image: "docker.io/busybox"
  command: ["sleep", "10000"]
```

# Executing our cutstom controller code locally
```bash
make
make manifests
make install
make run
```

# Create a sample CR
```bash
oc create -f ./samples
```

# Sources
- [kubebuilder book](https://book.kubebuilder.io/)
- [Programming kubernetes](https://learning.oreilly.com/library/view/programming-kubernetes/9781492047094/)
