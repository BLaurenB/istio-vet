# Mesh Version

The `meshversion` vetter helps detect mismatched, possibly incompatible versions
of [Istio](https://archive.istio.io/v0.8/docs/concepts/) components running in the mesh.

Vetter `meshversion` considers the version of Istio

[Mixer](https://archive.istio.io/v0.8/docs/concepts/policies-and-telemetry/overview/)
 image specified in the `istio-mixer` deployment as the *Istio version* for the cluster.

It compares the versions of other installed Istio components like
[Pilot](https://archive.istio.io/v0.8/docs/concepts/traffic-management/pilot/) 
with the *Istio version* and generates notes on version mismatch.

It also inspects the version of sidecar proxy deployed in pods in the mesh.
Notes are generated if any version differs from *Istio version*.

Version mismatch in various components can lead to unexpected behavior or policy
violations due to incompatibility. It is recommended to upgrade the reported
components to the *Istio version*.


-----------------------
The `meshversion` vetter helps detect mismatched, possibly incompatible versions
of [Istio](https://istio.io/docs/concepts/) components running in the mesh.

When automatic sidecar deployment is enabled for all pods in the mesh, this vetter compares the version of Istio in the installed version of Aspen Mesh to each pod in the mesh and generates notes upon version mismatch.

Vetter `meshversion` 
The istio-sidecar-injector ConfigMap has the sidecar & init images that will be injected into all new deployments, daemonsets, ....  If that doesn't match the images that are in existing Pods, emit a warning.
// GetInitializerConfig retrieves the Istio Initializer config.
// Istio Initializer config is stored as "istio-sidecar-injector" configmap in
// "istio-system" Namespace.

docker.io/istio/proxyv2:0.8.0
docker.io/istio/proxy_init:0.8.0


## Notes Generated

- [Mismatched sidecar version](README-sidecar-image-mismatch.md)
- [Mismatched init container version](README-init-image-mismatch.md)

