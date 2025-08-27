export default [
  { text: 'Module Configuration', link: './00-20-configure-docker-registry' },
  { text: 'Storage Configuration', link: './00-30-storage-configuration' },
  { text: 'Tutorials', link: './tutorials/README', collapsed: true, items: [
    { text: 'Use Docker Registry Internally', link: './tutorials/01-10-use-registry-internally' },
    { text: 'Expose Docker Registry', link: './tutorials/01-20-expose-registry' },
    { text: 'Remove Image Manifest', link: './tutorials/01-30-remove-image-manifest' }
  ]},
  { text: 'Resources', link: './resources/README', collapsed: true, items: [
    { text: 'Docker Registry Custom Resource', link: './resources/06-20-docker-registry-cr' }
  ]}
];
