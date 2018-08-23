# Sidecar Image Mismatch

## Example

The pod `your-app-45574414-qhgq3` in namespace `your-app` is running with
sidecar proxy image `docker.io/istio/proxyv2:1.0.0` but your environment is
injecting `docker.io/istio/proxyv2:0.8.0` for new workloads. Consider upgrading
the sidecar proxy in the pod.

## Description

The service mesh functions by injecting a sidecar proxy container into every
Kubernetes pod. Sidecars communicate with each other and with the control plane
to enable mesh features (such as automatic load-balancing, telemetry collection,
and authentication/authorization).

The `istio-sidecar-injector` configmap specifies which sidecar container Image
to inject into new workloads (like Deployments, DaemonSets, and Jobs) when they
are configured in your cluster.

This vetter checks pods in namespaces where automatic sidecar injection has been
enabled. The default setting is manual injection, but Aspen Mesh works best with
automatic injection. See [Managing Sidecar
Injection](https://my.aspenmesh.io/client/docs/getting-started/#managing-sidecar-injection)
for more information

A warning is generated when a pod is using a sidecar container from a different
image than what is specified in the `istio-sidecar-injector` configmap. If that
pod is deleted or crashes, the new pod would be injected with a sidecar matching
the image from the configmap. 

Mismatched images can be problematic for different reasons such as: 
- missing features, bugfixes, or security patches
- not compatible with other sidecars or the control plane
- different image from an unanticipated pod restart may cause unwanted behavior


## Suggested Resolution

Upgrade the sidecar image for these workloads to match the version in the
`istio-sidecar-injector` configmap, by doing one of the following:

- re-create (delete or roll out) this pod so it is injected with a new sidecar
  matching the version in the configmap
- edit the workload (for example, the Deployment)
