const execa = require('execa');
const fs = require('fs');
const {
  getEnvOrThrow,
  debug,
  wait,
} = require('../utils');
const {inspect} = require('util');

class KCPConfig {
  static fromEnv() {
    return new KCPConfig(
        getEnvOrThrow('KCP_KEB_API_URL'),
        getEnvOrThrow('KCP_OIDC_ISSUER_URL'),
        getEnvOrThrow('KCP_GARDENER_NAMESPACE'),
        getEnvOrThrow('KCP_TECH_USER_LOGIN'),
        getEnvOrThrow('KCP_TECH_USER_PASSWORD'),
        getEnvOrThrow('KCP_OIDC_CLIENT_ID'),
        getEnvOrThrow('KCP_OIDC_CLIENT_SECRET'),
        getEnvOrThrow('KCP_MOTHERSHIP_API_URL'),
        getEnvOrThrow('KCP_KUBECONFIG_API_URL'),
    );
  }
  constructor(host,
      issuerURL,
      gardenerNamespace,
      username,
      password,
      clientID,
      clientSecret,
      motherShipApiUrl,
      kubeConfigApiUrl) {
    this.host = host;
    this.issuerURL = issuerURL;
    this.gardenerNamespace = gardenerNamespace;
    this.username = username;
    this.password = password;
    this.clientID = clientID;
    this.clientSecret = clientSecret;
    this.motherShipApiUrl = motherShipApiUrl;
    this.kubeConfigApiUrl = kubeConfigApiUrl;
  }
}

class KCPWrapper {
  constructor(config) {
    this.kcpConfigPath = config.kcpConfigPath;
    this.gardenerNamespace = config.gardenerNamespace;
    this.clientID = config.clientID;
    this.clientSecret = config.clientSecret;
    this.issuerURL = config.issuerURL;
    this.motherShipApiUrl = config.motherShipApiUrl;
    this.kubeConfigApiUrl = config.kubeConfigApiUrl;

    this.username = config.username;
    this.password = config.password;
    this.host = config.host;

    this.kcpConfigPath = 'config.yaml';
    const stream = fs.createWriteStream(`${this.kcpConfigPath}`);
    stream.once('open', (_) => {
      stream.write(`gardener-namespace: ${this.gardenerNamespace}\n`);
      stream.write(`oidc-client-id: ${this.clientID}\n`);
      stream.write(`oidc-client-secret: ${this.clientSecret}\n`);
      stream.write(`keb-api-url: ${this.host}\n`);
      stream.write(`oidc-issuer-url: ${this.issuerURL}\n`);
      stream.write(`mothership-api-url: ${this.motherShipApiUrl}\n`);
      stream.write(`kubeconfig-api-url: ${this.kubeConfigApiUrl}\n`);
      stream.write(`username: ${this.username}\n`);
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
    const args = ['login', '-u', `${this.username}`, '-p', `${this.password}`];
    return await this.exec(args);
  }

  async version() {
    const args = ['--version'];
    return await this.exec(args);
  }

  async upgradeKyma(instanceID, kymaUpgradeVersion, upgradeTimeoutMin = 30) {
    const args = ['upgrade', 'kyma', `--version=${kymaUpgradeVersion}`, '--target', `instance-id=${instanceID}`];
    try {
      const res = await this.exec(args);

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
        const orchestrationStatus = await this.ensureOrchestrationSucceeded(orchestrationID, upgradeTimeoutMin);
        return orchestrationStatus;
      } catch (error) {
        debug(error);
      }

      try {
        const runtime = await this.runtimes({instanceID: instanceID});
        debug(`Runtime Status: ${inspect(runtime, false, null, false)}`);
      } catch (error) {
        debug(error);
      }

      try {
        const orchestration = await this.getOrchestrationStatus(orchestrationID);
        debug(`Orchestration Status: ${inspect(orchestration, false, null, false)}`);
      } catch (error) {
        debug(error);
      }

      try {
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
        const getReconciliationsInfo = await this.getReconciliationsInfo(i.schedulingID);
        console.log(`reconciliation info: ${i.schedulingID}: ${getReconciliationsInfo}`);
      }
    } catch {
      console.log('skipping reconciliation logging: error in reconcileInformationLog');
    }
  }


  async exec(args) {
    try {
      const defaultArgs = [
        '--config', `${this.kcpConfigPath}`,
      ];
      // debug([`>  kcp`, defaultArgs.concat(args).join(" ")].join(" "))
      const output = await execa('kcp', defaultArgs.concat(args));
      // debug(output);
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
