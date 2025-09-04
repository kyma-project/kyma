# Istio Controller RBAC Configuration

Security is paramount, so Istio Controller strictly follows the least privilege principle. While it needs permissions to manage Istio resources effectively, 
they're carefully tailored to specific tasks, avoiding unnecessary escalation to the level of all created resources.
As Istio Controller orchestrates the deployment of Istio components, it necessitates comprehensive management privileges for Istio resources. 
These privileges must mirror the access control levels accorded to the resources themselves, ensuring seamless operation.

## Elevated Permissions for ClusterRoles
Istio's installation grants the `istiod-clusterrole-istio-system` broad permissions by using `*` verbs for accessing the `ingresses/status` resource in the `networking.k8s.io` API group.
The ClusterRole of Istio Controller therefore also requires broad cluster role permissions (`*`).