const fs = require("fs");
const { join } = require("path");
var AdmZip = require("adm-zip");
const axios = require("axios");
const { debug } = require("../utils");


async function downloadFile(url, filename) {
  const writer = fs.createWriteStream(filename)

  const response = await axios.get(url, { responseType: 'stream' });

  response.data.pipe(writer)

  return new Promise((resolve, reject) => {
    writer.on('finish', resolve)
    writer.on('error', reject)
  })
}

/**
 * Downloads Kyma branch into temporary folder as zip, extracts charts, and returns full path to resources folder
 * @param {Object} options Optional parameters
 * @param {string} options.repo Github repository with Kyma charts, default: kyma-project/kyma
 * @param {string} options.source Branch that should be downloaded from repository, default: master
 * @return {string} Path to the resources folder
 */
async function downloadCharts(options) {
  options = options || {};
  const repo = options.repo || 'kyma-project/kyma'
  const branch = options.source || 'master';
  const dir = 'tmp-' + branch;
  const resourcesPath = join(dir, 'resources')
  debug("Resource path:",resourcesPath)
  if (fs.existsSync(resourcesPath)) {
    debug("The resources path already exists - download skipped")
    return;
  }
  if (!fs.existsSync(dir)){
    fs.mkdirSync(dir);
  }

  const zipFile = join(dir, branch + '.zip')
  const url = `https://codeload.github.com/${repo}/zip/${branch}`;
  debug("Downloading Kyma charts from ",url);
  await downloadFile(url, zipFile)
  debug("Kyma charts downloaded");
  var zip = new AdmZip(zipFile);
  var zipEntries = zip.getEntries();
  zipEntries.forEach(function (zipEntry) {
    const target = zipEntry.entryName.split('/').slice(1).join('/');
    if (target == 'resources/') {
      zip.extractEntryTo(zipEntry.entryName, dir);
      fs.renameSync(join(dir, zipEntry.entryName), resourcesPath);
    }
  });
  return resourcesPath;
}

module.exports = { downloadCharts };