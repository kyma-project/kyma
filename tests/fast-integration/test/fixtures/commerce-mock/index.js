const fs = require("fs");
const path = require("path");

const commerceMockYaml = fs.readFileSync(
  path.join(__dirname, "./commerce-mock.yaml"),
  {
    encoding: "utf8",
  }
);

const lastorderFunctionYaml = fs.readFileSync(
  path.join(__dirname, "./lastorder-function.yaml"),
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
  name: commerce-binding
spec:
  instanceRef:
    name: commerce-webservices`;


module.exports = {
  commerceMockYaml,
  serviceCatalogResources,
  mocksNamespaceYaml,
  lastorderFunctionYaml,
};
