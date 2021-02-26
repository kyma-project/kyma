#!/usr/bin/env node
const installer = require("./installer")
const { switchDebug } = require("./utils");

function installOptions(yargs) {
  yargs.options({
    'source': {
      alias: 's',
      describe: 'Installation source. \n\
          - To use a specific release, write "kyma install --source=1.15.1".\n\
          - To use the local sources, write "kyma install --source=local".'
    },
    'skip-components': {
      describe: 'Skip components from the list (comma separated)'
    },
    'upgrade': {
      describe: 'Upgrade components if already installed'
    },
    'new-eventing': {
      describe: 'Install new eventing instead of knative'
    }
  });

}
function uninstallOptions(yargs) {
  yargs.options({
    'skip-crd': {
      describe: 'Do not delete CRDs'
    },
    'skip-istio': {
      describe: 'Do not delete istio'
    }
  });

}

function verbose(argv) {
  if (argv.verbose) {
    switchDebug(true);
  }
}
const argv = require('yargs/yargs')(process.argv.slice(2))
  .usage('Usage: $0 <command> [options]')
  .options({ 'verbose': { alias: 'v', describe: 'Displays details of actions triggered by the command.' } })
  .command('install', 'Installs Kyma on a running Kubernetes cluster', installOptions, install)
  .command('uninstall', 'Removes Kyma from cluster', uninstallOptions, uninstall)
  .demandCommand(1, 1, 'Command is missing')
  .example('$0 install --skip-modules=monitoring,tracing,kiali')
  .strict()
  .wrap(null)
  .help()
  .argv

async function install(argv) {
  let src = undefined;
  verbose(argv);
  if (argv.source) {
    src = await installer.downloadCharts(argv)
  }
  const skipComponents = argv.skipComponents ? argv.skipComponents.split(',').map(c => c.trim()) : [];

  await installer.installKyma({ resourcesPath: src, skipComponents, isUpgrade: !!argv.upgrade });
  console.log('Kyma installed')
}

async function uninstall(argv) {
  verbose(argv);
  await installer.uninstallKyma(argv);
  console.log('Kyma uninstalled')
}

