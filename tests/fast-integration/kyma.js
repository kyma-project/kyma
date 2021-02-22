#!/usr/bin/env node
const installer = require("./installer")
const axios = require("axios");
var AdmZip = require('adm-zip');
var fs = require('fs');
const { join } = require("path");

function installOptions(yargs) {
  yargs.options({
    'source': {
      alias: 's',
      describe: 'Installation source. \n\
          - To use a specific release, write "kyma install --source=1.15.1".\n\
          - To use the local sources, write "kyma install --source=local".'
    },
    'skip-modules': {
      describe: 'Skip modules from the list (comma separated)'
    }
  });

}
function verbose(argv){
  if (argv.verbose) {
    process.env.DEBUG="true";
  }
}
const argv = require('yargs/yargs')(process.argv.slice(2))
  .usage('Usage: $0 <command> [options]')
  .options({'verbose':{alias:'v', describe:'Displays details of actions triggered by the command.'}})
  .command('install', 'Installs Kyma on a running Kubernetes cluster', installOptions, install)
  .command('uninstall', 'Removes Kyma from cluster', {}, uninstall)
  .command('download', 'Downloads Kyma sources into ./tmp folder', {}, download)
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
    await download(argv)
    src = join(__dirname, './tmp/resources')
  }
  await installer.installKyma(src, "1.8.2", "");
  console.log('Kyma installed')
}

async function uninstall(argv) {
  verbose(argv);
  await installer.uninstallKyma();
  console.log('Kyma uninstalled')
}


async function downloadFile(url, filename) {
  const writer = fs.createWriteStream(filename)

  const response = await axios.get(url, { responseType: 'stream' });

  response.data.pipe(writer)

  return new Promise((resolve, reject) => {
    writer.on('finish', resolve)
    writer.on('error', reject)
  })
}


async function download(argv) {
  verbose(argv);
  var dir = join(__dirname, './tmp');

  if (fs.existsSync(dir)) {
    fs.rmdirSync(dir, { recursive: true })
  }
  fs.mkdirSync(dir);

  const repo = 'kyma-project/kyma'
  const branch = argv.source || 'master';
  const zipFile = join(__dirname, './tmp/', branch + '.zip')
  await downloadFile(`https://codeload.github.com/${repo}/zip/${branch}`, zipFile)
  var zip = new AdmZip(zipFile);
  var zipEntries = zip.getEntries();
  zipEntries.forEach(function (zipEntry) {
    const target = zipEntry.entryName.split('/').slice(1).join('/');
    if (target == 'resources/') {
      zip.extractEntryTo(zipEntry.entryName, dir);
      fs.renameSync(join(dir, zipEntry.entryName), join(dir, 'resources'));
    }
  });
}