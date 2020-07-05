# manifest-capturer
![test](https://github.com/terakoya76/manifest-capturer/workflows/test/badge.svg)

## Why to use manifest-capturer
If you use managed kubernetes service such as EKS/GKE/AKE, you will find that your worker nodes have some `managed resources` such as coredns.

Most people manage all of their resources as IaC declaretive manifests, but to extend such `managed resources`, you need to edit these resources directly through the Kubernetes API.

The main motivation for manifest-capturer is to manage `managed resources` independently from an IaC perspective.

You can specify a target `managed resource` in the Capturer CR, and manifest-capturer will capture changes for that resource.

The captured change will be delivered to the destination specified in the Output CR (typically GitHub), and kept as a snapshot.

if the `managed resource` change has an unintended effect, you can use snapshots to investigate and recover from the failure.

## Quickstart
```bash
$ minikube start

# register CRD on schema
$ make install

# deploy sample CRs on your cluster
$ kubectl apply -f config/samples/capturer_v1alpha1_capturer/configmap_capturer.yaml -n kube-system
$ kubectl apply -f config/samples/capturer_v1alpha1_output/configmap_github_output.yaml -n kube-system

# run manifest-capturer controller
$ make run ENABLE_WEBHOOKS=false
```

## Supported
### Resource to be captured
* Deployment
* ConfigMap

### Output to be published
* GitHub
* Slack

## Examples
Check out the [config/sample](https://github.com/terakoya76/manifest-capturer/config/samples) directory to see some examples
