// until we have a proper config
const apiRuleName = 'fast-integration-test';
const domain = 'kyma-local';

const config = {
  functionUrl: `https://${apiRuleName}.${domain}`
};

module.exports = config;
