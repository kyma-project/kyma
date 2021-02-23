const {  getAllResourceTypes, listResources, deleteNamespaces } = require("../utils");


describe("K8S resource discovery", function () {
  this.timeout(3 * 60 * 1000);
  it.skip("should show resources", async function () {
    const types = await getAllResourceTypes()
    for (let type of types) {
      const path = `${type.path}/namespaces/default/${type.name}`;
      const result = await Promise.all(['mocks', 'test', 'orders-service'].map((n) => listResources(`${type.path}/namespaces/${n}/${type.name}`)));
      let names = result.flat()
      if (names.length) {
        console.log(type.path, type.name, ':', (names.length < 10) ? names : names.length + " items");
      }
    }
  })
  it("Should delete all resources in test namespaces",async function(){
    await deleteNamespaces(['mocks','orders-service','test'])
  })


})

