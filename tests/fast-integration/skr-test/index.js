module.exports = {
  ...require('./helpers'),
  ...require('./provision/provision-skr'),
  ...require('./oidc/index'),
  ...require('./machine-type/index'),
  ...require('./btp-manager-secret/index'),
};
