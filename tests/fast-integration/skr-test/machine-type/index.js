const {expect} = require('chai');
const {updateSKR, getCatalog} = require('../../kyma-environment-broker');
const {keb, kcp} = require('../helpers');
const {GardenerClient, GardenerConfig} = require('../../gardener');
const {getEnvOrThrow} = require('../../utils');

const updateTimeout = 1000 * 60 * 30; // 30m

function machineTypeE2ETest(getShootOptionsFunc, getShootInfoFunc) {
  describe('Machine type update test', function() {
    let shoot = undefined;
    let options = undefined;
    let updateMachineType = undefined;
    let defaultMachineType = undefined;
    const planID = getEnvOrThrow('KEB_PLAN_ID');
    const gardener = new GardenerClient(GardenerConfig.fromEnv());

    before('Get provisioned Shoot Info', async function() {
      shoot = getShootInfoFunc();
      options = getShootOptionsFunc();
    });

    it('Should check default machine type after provisioning', async function() {
      defaultMachineType = await getMachineType(gardener, shoot.name);
      console.log(`Default machine type ${defaultMachineType}`);
    });

    it('Should figure out next machine type to update to', async function() {
      const catalog = await getCatalog(keb);
      console.log(`catalog services ${catalog.services.length}`);
      for (let i = 0; i < catalog.services.length; i++) {
        const s = catalog.services[i];
        console.log(`checking catalog service ${s.id}`);
        if (s.id !== '47c9dcbf-ff30-448e-ab36-d3bad66ba281') {
          continue;
        }
        for (let j = 0; j < s.plans.length; j++) {
          const p = s.plans[j];
          console.log(`checking catalog plan ${p.id}, looking for ${planID}`);
          if (p.id !== planID) {
            continue;
          }
          const si = p.schemas.service_instance;
          console.log(`catalog plan update ${si.update !== undefined}`);
          if (si.update === undefined || si.update.parameters.properties.machineType === undefined) {
            continue;
          }
          const mt = si.update.parameters.properties.machineType.enum;
          console.log(`catalog plan update machine types ${mt}`);
          if (mt.length < 2) {
            continue;
          }
          for (let i = 0; i < mt.length; i++) {
            console.log(`checking catalog plan update machine type comp ${mt[i]}, looking for ${defaultMachineType}`);
            if (mt[i] === defaultMachineType) {
              console.log(`catalog plan update machine type ${mt[i]} === ${defaultMachineType}`);
              updateMachineType = mt[(i+1) % mt.length];
              break;
            }
          }
          if (updateMachineType !== undefined) {
            break;
          }
        }
      }
      console.log(`determined machine type to update: ${updateMachineType}`);
    });

    it('Should update SKR service instance machine type', async function() {
      if (updateMachineType === undefined) {
        console.log('skipping machine type update');
        return;
      }
      console.log(`updating ${defaultMachineType} => ${updateMachineType}`);
      this.timeout(updateTimeout);
      const customParams = {
        machineType: updateMachineType,
      };
      const skr = await updateSKR(keb,
          kcp,
          gardener,
          options.instanceID,
          shoot.name,
          customParams,
          updateTimeout,
          null,
          false);
      shoot = skr.shoot;
    });

    it('Should check machine type after update', async function() {
      if (updateMachineType === undefined) {
        console.log('skipping machine type update');
        return;
      }
      const machineType = await getMachineType(gardener, shoot.name);
      expect(machineType).to.equal(updateMachineType);
    });
  });
}

async function getMachineType(gardener, shoot) {
  await gardener.waitForShoot(shoot, 'Reconcile');
  const sh = await gardener.getShoot(shoot);
  return sh.spec.provider.workers[0].machine.type;
}

module.exports = {
  machineTypeE2ETest,
};
