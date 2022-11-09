const execa = require('execa');
const {getEnvOrThrow} = require('../utils');

async function smInstanceBinding(creds, btpOperatorInstance, btpOperatorBinding) {
  let args = [];
  try {
    args = ['login',
      '-a',
      creds.url,
      '--param',
      'subdomain=e2etestingscmigration',
      '--auth-flow',
      'client-credentials'];
    await execa('smctl', args.concat(['--client-id', creds.clientid, '--client-secret', creds.clientsecret]));

    args = ['provision', btpOperatorInstance, 'service-manager', 'service-operator-access', '--mode=sync'];
    await execa('smctl', args);

    // Move to Operator Install
    args = ['bind', btpOperatorInstance, btpOperatorBinding, '--mode=sync'];
    await execa('smctl', args);

    args = ['get-binding', btpOperatorBinding, '-o', 'json'];
    const out = await execa('smctl', args);
    const b = JSON.parse(out.stdout);
    const c = b.items[0].credentials;

    return {
      clientId: c.clientid,
      clientSecret: c.clientsecret,
      smURL: c.sm_url,
      url: c.url,
      instanceId: b.items[0].service_instance_id,
    };
  } catch (error) {
    if (error.stderr === undefined) {
      throw new Error(`failed to process output of "smctl ${args.join(' ')}"`);
    }
    throw new Error(`failed "smctl ${args.join(' ')}": ${error.stderr}`);
  }
}

async function provisionPlatform(creds, svcatPlatform) {
  let args = [];
  try {
    args = ['login',
      '-a',
      creds.url,
      '--param',
      'subdomain=e2etestingscmigration',
      '--auth-flow',
      'client-credentials'];
    await execa('smctl', args.concat(['--client-id', creds.clientid, '--client-secret', creds.clientsecret]));

    // $ smctl register-platform <name> kubernetes -o json
    // Output:
    // {
    //   "id": "<platform-id/cluster-id>",
    //   "name": "<name>",
    //   "type": "kubernetes",
    //   "created_at": "...",
    //   "updated_at": "...",
    //   "credentials": {
    //     "basic": {
    //       "username": "...",
    //       "password": "..."
    //     }
    //   },
    //   "labels": {
    //     "subaccount_id": [
    //       "..."
    //     ]
    //   },
    //   "ready": true
    // }
    args = ['register-platform', svcatPlatform, 'kubernetes', '-o', 'json'];
    const registerPlatformOut = await execa('smctl', args);
    const platform = JSON.parse(registerPlatformOut.stdout);

    return {
      clusterId: platform.id,
      name: platform.name,
      credentials: platform.credentials.basic,
    };
  } catch (error) {
    if (error.stderr === undefined) {
      throw new Error(`failed to process output of "smctl ${args.join(' ')}"`);
    }
    throw new Error(`failed "smctl ${args.join(' ')}": ${error.stderr}`);
  }
}

async function markForMigration(creds, svcatPlatform, btpOperatorInstanceId) {
  let errors = [];
  let args = [];
  try {
    args = ['login',
      '-a',
      creds.url,
      '--param',
      'subdomain=e2etestingscmigration',
      '--auth-flow',
      'client-credentials'];
    await execa('smctl', args.concat(['--client-id', creds.clientid, '--client-secret', creds.clientsecret]));
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n`]);
  }

  try {
    // usage: smctl curl -X PUT -d '{"sourcePlatformID": ":platformID"}' /v1/migrate/service_operator/:instanceID
    const data = {sourcePlatformID: svcatPlatform};
    args = ['curl', '-X', 'PUT', '-d', JSON.stringify(data), '/v1/migrate/service_operator/' + btpOperatorInstanceId];
    await execa('smctl', args);
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n`]);
  }
  if (errors.length > 0) {
    throw new Error(errors.join(', '));
  }
}

async function cleanupInstanceBinding(creds, svcatPlatform, btpOperatorInstance, btpOperatorBinding) {
  let errors = [];
  let args = [];
  try {
    args = ['login',
      '-a',
      creds.url,
      '--param',
      'subdomain=e2etestingscmigration',
      '--auth-flow',
      'client-credentials'];
    await execa('smctl', args.concat(['--client-id', creds.clientid, '--client-secret', creds.clientsecret]));
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n`]);
  }

  try {
    args = ['unbind', btpOperatorInstance, btpOperatorBinding, '-f', '--mode=sync'];
    const {stdout} = await execa('smctl', args);
    if (stdout !== 'Service Binding successfully deleted.') {
      errors = errors.concat([`failed "smctl ${args.join(' ')}": ${stdout}`]);
    }
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
  }

  try {
    // hint: probably should fail cause that instance created other instannces (after the migration is done)
    args = ['deprovision', btpOperatorInstance, '-f', '--mode=sync'];
    const {stdout} = await execa('smctl', args);
    if (stdout !== 'Service Instance successfully deleted.') {
      errors = errors.concat([`failed "smctl ${args.join(' ')}": ${stdout}`]);
    }
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
  }

  try {
    args = ['delete-platform', svcatPlatform, '-f', '--cascade'];
    await execa('smctl', args);
    // if (stdout !== "Platform(s) successfully deleted.") {
    //     errors = errors.concat([`failed "smctl ${args.join(' ')}": ${stdout}`])
    // }
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n`]);
  }

  if (errors.length > 0) {
    throw new Error(errors.join(', '));
  }
}


class AdminCreds {
  static fromEnv() {
    return new AdminCreds(
        getEnvOrThrow('BTP_SM_ADMIN_CLIENTID'),
        getEnvOrThrow('BTP_SM_ADMIN_CLIENTSECRET'),
        getEnvOrThrow('BTP_SM_ADMIN_URL'),
    );
  }

  constructor(clientid, clientsecret, url) {
    this.clientid = clientid;
    this.clientsecret = clientsecret;
    this.url = url;
  }
}

class BTPOperatorCreds {
  static fromEnv() {
    return new BTPOperatorCreds(
        getEnvOrThrow('BTP_OPERATOR_CLIENTID'),
        getEnvOrThrow('BTP_OPERATOR_CLIENTSECRET'),
        getEnvOrThrow('BTP_OPERATOR_URL'),
        getEnvOrThrow('BTP_OPERATOR_TOKENURL'),
    );
  }

  static dummy() {
    return new BTPOperatorCreds(
        'dummy_client_id',
        'dummy_client_secret',
        'dummy_url',
        'dummy_token_url',
    );
  }

  constructor(clientid, clientsecret, smURL, url) {
    this.clientid = clientid;
    this.clientsecret = clientsecret;
    this.smURL = smURL;
    this.url = url;
  }
}

module.exports = {
  provisionPlatform,
  smInstanceBinding,
  markForMigration,
  cleanupInstanceBinding,
  AdminCreds: AdminCreds,
  BTPOperatorCreds: BTPOperatorCreds,
};
