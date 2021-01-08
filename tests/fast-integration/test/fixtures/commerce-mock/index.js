const fs = require("fs");
const path = require("path");

const commerceMockYaml = fs.readFileSync(
  path.join(__dirname, "./commerce-mock.yaml"),
  {
    encoding: "utf8",
  }
);

const mocksNamespaceYaml = fs.readFileSync(
  path.join(__dirname, "./mocks-namespace.yaml"),
  {
    encoding: "utf8",
  }
);

const genericServiceClass = (name, namespace) => `
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceClass
metadata:
  name: ${name}
  namespace: ${namespace}
`;

const serviceCatalogResources = (
  webServicesExternalName,
  eventsExternalName
) => `
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: commerce-webservices
spec:
  serviceClassExternalName: ${webServicesExternalName}
---
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: commerce-events
spec:
  serviceClassExternalName: ${eventsExternalName}
---
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  labels:
    function: lastorder
  name: commerce-lastorder-binding
spec:
  instanceRef:
    name: commerce-webservices
---
apiVersion: servicecatalog.kyma-project.io/v1alpha1
kind: ServiceBindingUsage
metadata:
  labels:
    function: lastorder
    serviceBinding: commerce-lastorder-binding
  name: commerce-lastorder-sbu
spec:
  serviceBindingRef:
    name: commerce-lastorder-binding
  usedBy:
    kind: serverless-function
    name: lastorder`;

const tokenRequestYaml = `apiVersion: applicationconnector.kyma-project.io/v1alpha1	
kind: TokenRequest	
metadata:	
  name: commerce`;

module.exports = {
  commerceMockYaml,
  genericServiceClass,
  serviceCatalogResources,
  mocksNamespaceYaml,
  tokenRequestYaml,
};
