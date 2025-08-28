import commandsSidebar from './gen-docs/_sidebar'
export default [
  { text: 'Tutorials', link: './tutorials/README', collapsed: true, items: [
    { text: 'Adding and Deleting a Kyma Module Using Kyma CLI', link: './tutorials/01-10-add-delete-modules' },
    { text: 'Setting Your Module to the Managed and Unmanaged State in Kyma CLI', link: './tutorials/01-11-manage-unmanage-modules' },
    { text: 'Running an Application Using the `app push` Command', link: './tutorials/01-20-app-push-command-usage' }
    ] },
  { text: 'Commands', link: './gen-docs/README', collapsed: true, items: commandsSidebar }
];
