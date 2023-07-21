const execa = require('execa');
const fs = require('fs');
const stream = require('stream');
const {
  getEnvOrThrow,
  debug,
  wait,
} = require('../../utils');
const {inspect} = require('util');

class KCPConfig {
  static fromEnv() {
    return new KCPConfig();
  }
  constructor() {
    this.host = getEnvOrThrow('KCP_KEB_API_URL');
    this.issuerURL = getEnvOrThrow('KCP_OIDC_ISSUER_URL');
    this.gardenerNamespace = getEnvOrThrow('KCP_GARDENER_NAMESPACE');
    this.username = getEnvOrThrow('KCP_TECH_USER_LOGIN');
    this.password = getEnvOrThrow('KCP_TECH_USER_PASSWORD');
    this.clientID = getEnvOrThrow('KCP_OIDC_CLIENT_ID');

    if (process.env.KCP_OIDC_CLIENT_SECRET) {
      this.clientSecret = getEnvOrThrow('KCP_OIDC_CLIENT_SECRET');
    } else {
      this.oauthClientID = getEnvOrThrow('KCP_OAUTH2_CLIENT_ID');
      this.oauthSecret = getEnvOrThrow('KCP_OAUTH2_CLIENT_SECRET');
      this.oauthIssuer = getEnvOrThrow('KCP_OAUTH2_ISSUER_URL');
    }

    this.motherShipApiUrl = getEnvOrThrow('KCP_MOTHERSHIP_API_URL');
    this.kubeConfigApiUrl = getEnvOrThrow('KCP_KUBECONFIG_API_URL');
  }
}

class KCPWrapper {
  constructor(config) {
    this.kcpConfigPath = config.kcpConfigPath;
    this.gardenerNamespace = config.gardenerNamespace;
    this.clientID = config.clientID;
    this.clientSecret = config.clientSecret;
    this.oauthClientID = config.oauthClientID;
    this.oauthSecret = config.oauthSecret;
    this.oauthIssuer = config.oauthIssuer;

    this.issuerURL = config.issuerURL;
    this.motherShipApiUrl = config.motherShipApiUrl;
    this.kubeConfigApiUrl = config.kubeConfigApiUrl;

    this.username = config.username;
    this.password = config.password;
    this.host = config.host;

    this.kcpConfigPath = 'config.yaml';
    const stream = fs.createWriteStream(`${this.kcpConfigPath}`);
    stream.once('open', (_) => {
      stream.write(`gardener-namespace: "${this.gardenerNamespace}"\n`);
      if (process.env.KCP_OIDC_CLIENT_SECRET) {
        stream.write(`oidc-client-id: "${this.clientID}"\n`);
        stream.write(`oidc-client-secret: ${this.clientSecret}\n`);
        stream.write(`username: ${this.username}\n`);
      } else {
        stream.write(`oauth2-client-id: "${this.oauthClientID}"\n`);
        stream.write(`oauth2-client-secret: "${this.oauthSecret}"\n`);
        stream.write(`oauth2-issuer-url: "${this.oauthIssuer}"\n`);
      }

      stream.write(`keb-api-url: "${this.host}"\n`);
      stream.write(`oidc-issuer-url: "${this.issuerURL}"\n`);
      stream.write(`mothership-api-url: "${this.motherShipApiUrl}"\n`);
      stream.write(`kubeconfig-api-url: "${this.kubeConfigApiUrl}"\n`);
      stream.end();
    });
  }

  async runtimes(query) {
    let args = ['runtimes', '--output', 'json'];
    if (query.account) {
      args = args.concat('--account', `${query.account}`);
    }
    if (query.subaccount) {
      args = args.concat('--subaccount', `${query.subaccount}`);
    }
    if (query.instanceID) {
      args = args.concat('--instance-id', `${query.instanceID}`);
    }
    if (query.runtimeID) {
      args = args.concat('--runtime-id', `${query.runtimeID}`);
    }
    if (query.region) {
      args = args.concat('--region', `${query.region}`);
    }
    if (query.shoot) {
      args = args.concat('--shoot', `${query.shoot}`);
    }
    if (query.state) {
      args = args.concat('--state', `${query.state}`);
    }
    if (query.ops) {
      args = args.concat('--ops');
    }
    const result = await this.exec(args);
    return JSON.parse(result);
  }

  async reconciliations(query) {
    let args = ['reconciliations', `${query.parameter}`, '--output', 'json'];
    if (query.runtimeID) {
      args = args.concat('--runtime-id', `${query.runtimeID}`);
    }
    if (query.schedulingID) {
      args = args.concat('--scheduling-id', `${query.schedulingID}`);
    }
    const result = await this.exec(args);
    return JSON.parse(result);
  }

  async login() {
    let args;
    if (process.env.KCP_OIDC_CLIENT_SECRET) {
      args = ['login', '-u', `${this.username}`, '-p', `${this.password}`];
    } else {
      args = ['login'];
    }

    return await this.exec(args);
  }

  async version() {
    const args = ['--version'];
    return await this.exec(args);
  }

  async upgradeKyma(instanceID, kymaUpgradeVersion, upgradeTimeoutMin = 30) {
    const args = ['upgrade', 'kyma', `--version=${kymaUpgradeVersion}`, '--target', `instance-id=${instanceID}`];
    try {
      console.log('Executing kcp upgrade');
      const res = await this.exec(args, true, true);

      console.log('Checking orchestration');
      // output if successful:
      // "Note: Ignore sending slack notification when slackAPIURL is empty\n" +
      // "OrchestrationID: 22f19856-679b-4e68-b533-f1a0a46b1eed"
      // so we need to extract the uuid
      if (!res.includes('OrchestrationID: ')) {
        throw new Error(`Kyma Upgrade failed. KCP upgrade command returned no OrchestrationID. Response: \"${res}\"`);
      }
      const orchestrationID = res.split('OrchestrationID: ')[1];
      debug(`OrchestrationID: ${orchestrationID}`);

      try {
        console.log('Ensure execution suceeded');
        const orchestrationStatus = await this.ensureOrchestrationSucceeded(orchestrationID, upgradeTimeoutMin);
        return orchestrationStatus;
      } catch (error) {
        debug(error);
      }

      try {
        console.log('Check runtime status');
        const runtime = await this.runtimes({instanceID: instanceID});
        debug(`Runtime Status: ${inspect(runtime, false, null, false)}`);
      } catch (error) {
        debug(error);
      }

      try {
        console.log('Check orchestration');
        const orchestration = await this.getOrchestrationStatus(orchestrationID);
        debug(`Orchestration Status: ${inspect(orchestration, false, null, false)}`);
      } catch (error) {
        debug(error);
      }

      try {
        console.log('Check operations');
        const operations = await this.getOrchestrationsOperations(orchestrationID);
        debug(`Operations: ${inspect(operations, false, null, false)}`);
      } catch (error) {
        debug(error);
      }

      throw new Error('Kyma Upgrade failed');
    } catch (error) {
      debug(error);
      throw new Error('failed during upgradeKyma');
    }
  };

  async getReconciliationsOperations(runtimeID) {
    await this.login();
    const reconciliationsOperations = await this.reconciliations({parameter: 'operations',
      runtimeID: runtimeID});
    return JSON.stringify(reconciliationsOperations, null, '\t');
  }

  async getReconciliationsInfo(schedulingID) {
    await this.login();
    const reconciliationsInfo = await this.reconciliations({parameter: 'info',
      schedulingID: schedulingID});

    return JSON.stringify(reconciliationsInfo, null, '\t');
  }

  async getRuntimeEvents(instanceID) {
    await this.login();
    return this.exec(['runtimes', '--instance-id', instanceID, '--events']);
  }

  async getRuntimeStatusOperations(instanceID) {
    await this.login();
    const runtimeStatus = await this.runtimes({instanceID: instanceID, ops: true});

    return JSON.stringify(runtimeStatus, null, '\t');
  }

  async getOrchestrationsOperations(orchestrationID) {
    // debug('Running getOrchestrationsOperations...')
    const args = ['orchestration', `${orchestrationID}`, 'operations', '-o', 'json'];
    try {
      const res = await this.exec(args);
      const operations = JSON.parse(res);
      // debug(`getOrchestrationsOperations output: ${operations}`)

      return operations;
    } catch (error) {
      debug(error);
      throw new Error('failed during getOrchestrationsOperations');
    }
  }

  async getOrchestrationsOperationStatus(orchestrationID, operationID) {
    // debug('Running getOrchestrationsOperationStatus...')
    const args = ['orchestration', `${orchestrationID}`, '--operation', `${operationID}`, '-o', 'json'];
    try {
      let res = await this.exec(args);
      res = JSON.parse(res);

      return res;
    } catch (error) {
      debug(error);
      throw new Error('failed during getOrchestrationsOperationStatus');
    }
  }

  async getOrchestrationStatus(orchestrationID) {
    // debug('Running getOrchestrationStatus...')
    const args = ['orchestrations', `${orchestrationID}`, '-o', 'json'];
    try {
      const res = await this.exec(args);
      const o = JSON.parse(res);

      debug(`OrchestrationID: ${o.orchestrationID} (${o.type} to version ${o.parameters.kyma.version})
      Status: ${o.state}`);

      const operations = await this.getOrchestrationsOperations(o.orchestrationID);
      // debug(`Got ${operations.length} operations for OrchestrationID ${o.orchestrationID}`)

      let upgradeOperation = {};
      if (operations.count > 0) {
        upgradeOperation = await this.getOrchestrationsOperationStatus(orchestrationID, operations.data[0].operationID);
        debug(`OrchestrationID: ${orchestrationID}
        OperationID: ${operations.data[0].operationID}
        OperationStatus: ${upgradeOperation[0].state}`);
      } else {
        debug(`No operations in OrchestrationID ${o.orchestrationID}`);
      }

      return o;
    } catch (error) {
      debug(error);
      throw new Error('failed during getOrchestrationStatus');
    }
  };

  async ensureOrchestrationSucceeded(orchenstrationID, upgradeTimeoutMin = 30) {
    // Decides whether to go to the next step of while or not based on
    // the orchestration result (0 = succeeded, 1 = failed, 2 = cancelled, 3 = pending/other)
    debug(`Waiting for Kyma Upgrade with OrchestrationID ${orchenstrationID} to succeed...`);
    try {
      const res = await wait(
          () => this.getOrchestrationStatus(orchenstrationID),
          (res) => res && res.state && (res.state === 'succeeded' || res.state === 'failed'),
          1000 * 60 * upgradeTimeoutMin, // 30 min
          1000 * 30, // 30 seconds
      );

      if (res.state !== 'succeeded') {
        debug('KEB Orchestration Status:', res);
        throw new Error(`orchestration didn't succeed in 15min: ${JSON.stringify(res)}`);
      }

      const descSplit = res.description.split(' ');
      if (descSplit[1] !== '1') {
        throw new Error(`orchestration didn't succeed (number of scheduled operations should be equal to 1):
        ${JSON.stringify(res)}`);
      }

      return res;
    } catch (error) {
      debug(error);
      throw new Error('failed during ensureOrchestrationSucceeded');
    }
  }

  async reconcileInformationLog(runtimeStatus) {
    try {
      const objRuntimeStatus = JSON.parse(runtimeStatus);

      try {
        if (!objRuntimeStatus.data[0].runtimeID) {}
      } catch (e) {
        console.log('skipping reconciliation logging: no runtimeID provided by runtimeStatus');
        return;
      }

      // kcp reconciliations operations -r <runtimeID> -o json
      const reconciliationsOperations = await this.getReconciliationsOperations(objRuntimeStatus.data[0].runtimeID);

      const objReconciliationsOperations = JSON.parse(reconciliationsOperations);

      if ( objReconciliationsOperations == null ) {
        console.log(`skipping reconciliation logging: kcp rc operations -r ${objRuntimeStatus.data[0].runtimeID}
         -o json returned null`);
        return;
      }

      const objReconciliationsOperationsLength = objReconciliationsOperations.length;

      if (objReconciliationsOperationsLength === 0) {
        console.log(`no reconciliation operations found`);
        return;
      }
      console.log(`number of reconciliation operations: ${objReconciliationsOperationsLength}`);

      // using only last three operations
      const lastObjReconciliationsOperations = objReconciliationsOperations.
          slice(Math.max(0, objReconciliationsOperations.length - 3), objReconciliationsOperations.length);

      for (const i of lastObjReconciliationsOperations) {
        console.log(`reconciliation operation status: ${i.status}`);

        // kcp reconciliations info -i <scheduling-id> -o json
        await this.getReconciliationsInfo(i.schedulingID);
      }
    } catch {
      console.log('skipping reconciliation logging: error in reconcileInformationLog');
    }
  }

  async exec(args, pipeStdout = false, sendYes = false) {
    try {
      const defaultArgs = [
        '--config', `${this.kcpConfigPath}`,
      ];
      // debug([`>  kcp`, defaultArgs.concat(args).join(" ")].join(" "))
      const subprocess = execa('kcp', defaultArgs.concat(args), {stdio: 'pipe'});

      if ( pipeStdout ) {
        subprocess.stdout.pipe(process.stdout);
      }

      if ( sendYes ) {
        const inStream = new stream.Readable();
        inStream.push('Y');
        inStream.push(null);
        inStream.pipe(subprocess.stdin);
      }

      const output = await subprocess;
      return output.stdout;
    } catch (err) {
      if (err.stderr === undefined) {
        throw new Error(`failed to process kcp binary output: ${err.toString()}`);
      }
      throw new Error(`kcp command failed: ${err.stderr.toString()}`);
    }
  }
}

module.exports = {
  KCPConfig,
  KCPWrapper,
};
