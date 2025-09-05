export default [
  { text: 'Create the <code>sap-btp-manager</code> Secret', link: './03-00-create-btp-manager-secret' },
  { text: 'Install the SAP BTP Operator Module', link: './03-05-install-module' },
  { text: 'Preconfigured Credentials and Access', link: './03-10-preconfigured-secret' },
  { text: 'Create Service Instances and Service Bindings', link: './03-30-create-instances-and-bindings' },
  { text: 'Update Service Instances', link: './03-31-update-service-instances' },
  { text: 'Delete Service Bindigs and Service Instances', link: './03-32-delete-bindings-and-instances' },
  { text: 'Migrate Service Instances and Service Bindings from a Custom SAP BTP Service Operator', link: './03-33-migrate-instances-and-bindings' },
  { text: 'Working with Multiple Subaccounts', link: './03-20-multitenancy' },
  { text: 'Instance-Level Mapping', link: './03-21-instance-level-mapping' },
  { text: 'Namespace-Level Mapping', link: './03-22-namespace-level-mapping' },
  { text: 'Pass Parameters', link: './03-60-pass-parameters' },
  { text: 'Rotate Service Bindings', link: './03-40-service-binding-rotation' },
  { text: 'Formats of Service Binding Secrets', link: './03-50-formatting-service-binding-secret' },
  { text: 'Resources', link: './resources/README', collapsed: true, items: [
    { text: 'SAP BTP Operator Custom Resource', link: './resources/02-10-sap-btp-operator-cr' },
    { text: 'Service Instance Custom Resource', link: './resources/02-20-service-instance-cr' },
    { text: 'Service Binding Custom Resource', link: './resources/02-30-service-binding-cr' }
    ] },
  { text: 'Tutorials', link: './tutorials/README', collapsed: true, items: [
    { text: 'Create an SAP BTP Service in Your Kyma Cluster', link: './tutorials/04-40-create-service-in-cluster' }
    ] }
];
