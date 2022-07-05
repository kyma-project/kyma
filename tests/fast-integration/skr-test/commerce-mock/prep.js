// const {
//   commerceMockTestPreparation,
//   gatherOptions,
//   withSuffix,
// } = require('../index');
// const {getEnvOrThrow} = require('../../utils');
// const {prepareCompassResources, initK8sConfig, getSKRConfig} = require('../helpers');
//
// describe('Preparations for CommerceMock tests', async function() {
//   const suffix = getEnvOrThrow('TEST_SUFFIX');
//   const instanceID = getEnvOrThrow('INSTANCE_ID');
//   const options = gatherOptions(
//       withSuffix(suffix),
//   );
//   let shoot;
//
//   context('Gather information from externally provisioned SKR', async function() {
//     it('Fetch SKR config', async function() {
//       shoot = await getSKRConfig(instanceID);
//     });
//
//     it('Prepare compass scenario', async function() {
//       await prepareCompassResources(shoot, options);
//     });
//
//     it('Initialize the k8s client', async function() {
//       await initK8sConfig(shoot);
//     });
//   });
//
//   context('Prepare the Commerce Mock resources', async function() {
//     commerceMockTestPreparation(options);
//   });
// });
