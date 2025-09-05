export default [
  { text: 'From Code to Function', link: './00-10-from-code-to-function' },
  { text: 'Configuring Serverless', link: './00-20-configure-serverless' },
  { text: 'Development Toolkit', link: './00-30-development-toolkit' },
  { text: 'Function Security', link: './00-40-security-considerations' },
  { text: 'Limitations', link: './00-50-limitations' },
  { text: 'Tutorials', link: './tutorials/README', collapsed: true, items: [
    { text: 'Create an Inline Function', link: './tutorials/01-10-create-inline-function' },
    { text: 'Create a Git Function', link: './tutorials/01-11-create-git-function' },
    { text: 'Expose the Function', link: './tutorials/01-20-expose-function' },
    { text: 'Manage Functions Through Kyma CLI', link: './tutorials/01-30-manage-functions-with-kyma-cli' },
    { text: 'Send and Receive Cloud Events', link: './tutorials/01-90-set-asynchronous-connection' },
    { text: 'Customize Function Traces', link: './tutorials/01-100-customize-function-traces' },
    { text: 'Override Runtime Image', link: './tutorials/01-110-override-runtime-image' },
    { text: 'Inject Environment Variables', link: './tutorials/01-120-inject-envs' },
    { text: 'Use External Scalers', link: './tutorials/01-130-use-external-scalers' },
    { text: 'Access to Secrets Mounted as Volume', link: './tutorials/01-140-use-secret-mounts' }
    ] },
  { text: 'Resources', link: './resources/README', collapsed: true, items: [
    { text: 'Function CR', link: './resources/06-10-function-cr' },
    { text: 'Serverless CR', link: './resources/06-20-serverless-cr' }
    ] },
  { text: 'Technical Reference', link: './technical-reference/README', collapsed: true, items: [
    { text: 'Buildless Serverless', link: './technical-reference/03-10-buildless-serverless' },
    { text: 'Serverless Architecture', link: './technical-reference/04-10-architecture' },
    { text: 'Internal Docker Registry', link: './technical-reference/04-20-internal-registry' },
    { text: 'Environment Variables in Functions', link: './technical-reference/05-20-env-variables' },
    { text: 'Sample Functions', link: './technical-reference/07-10-sample-functions' },
    { text: 'Function Processing', link: './technical-reference/07-20-function-processing-stages' },
    { text: 'Git Source Type', link: './technical-reference/07-40-git-source-type' },
    { text: 'Function\'s Specification', link: './technical-reference/07-70-function-specification' },
    { text: 'Available Presets', link: './technical-reference/07-80-available-presets' },
    { text: 'Serverless Buildless Mode', link: './technical-reference/03-10-buildless-serverless' }
    ] },
  { text: 'Troubleshooting Guides', link: './troubleshooting-guides/README', collapsed: true, items: [
    { text: 'Functions Won\'t Build', link: './troubleshooting-guides/03-10-cannot-build-functions' },
    { text: 'Container Fails', link: './troubleshooting-guides/03-20-failing-function-container' },
    { text: 'Functions Failing To Build on k3d', link: './troubleshooting-guides/03-40-function-build-failing-k3d' },
    { text: 'Serverless Periodically Restarting', link: './troubleshooting-guides/03-50-serverless-periodically-restaring' }
    ] },
  { text: 'Best Practices', link: './08-10-best-practices' }
];
