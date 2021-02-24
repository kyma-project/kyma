const fs = require("fs");
const { join } = require("path");
var AdmZip = require("adm-zip");
const axios = require("axios");


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
  var dir = fs.mkdtempSync('kyma');
  const resourcesPath = join(dir, 'resources')

  const zipFile = join(dir, branch + '.zip')
  await downloadFile(`https://codeload.github.com/${repo}/zip/${branch}`, zipFile)
  var zip = new AdmZip(zipFile);
  var zipEntries = zip.getEntries();
  zipEntries.forEach(function (zipEntry) {
    const target = zipEntry.entryName.split('/').slice(1).join('/');
    if (target == 'resources/') {
      zip.extractEntryTo(zipEntry.entryName, dir);
      fs.renameSync(join(dir, zipEntry.entryName), resourcesPath );
    }
  });
  return resourcesPath;
}

module.exports = {downloadCharts};