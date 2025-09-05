export default [
  { text: 'NFS', link: './00-20-nfs' },
  { text: 'VPC Peering', link: './00-30-vpc-peering' },
  { text: 'Authorizing Cloud Manager in the Remote Cloud Provider', link: './00-31-vpc-peering-authorization' },
  { text: 'Redis', link: './00-40-redis' },
  { text: 'Resources', link: './resources/README' },
  { text: 'Tutorials', link: './tutorials/README', collapsed: true, items: [
    { text: 'Using NFS in Amazon Web Services', link: './tutorials/01-20-10-aws-nfs-volume' },
    { text: 'Using NFS in Google Cloud', link: './tutorials/01-20-20-gcp-nfs-volume' },
    { text: 'Creating VPC Peering in Amazon Web Services', link: './tutorials/01-30-10-aws-vpc-peering' },
    { text: 'Creating VPC Peering in Google Cloud', link: './tutorials/01-30-20-gcp-vpc-peering' },
    { text: 'Creating VPC Peering in Microsoft Azure', link: './tutorials/01-30-30-azure-vpc-peering' },
    { text: 'Using AwsRedisInstance Custom Resources', link: './tutorials/01-40-10-aws-redis-instance' },
    { text: 'Using GcpRedisInstance Custom Resources', link: './tutorials/01-40-20-gcp-redis-instance' },
    { text: 'Using AzureRedisInstance Custom Resources', link: './tutorials/01-40-30-azure-redis-instance' },
    { text: 'Using AzureRedisCluster Custom Resources', link: './tutorials/01-50-30-azure-redis-cluster' }
    ] },
  { text: 'Glossary', link: './00-10-glossary' }
];
