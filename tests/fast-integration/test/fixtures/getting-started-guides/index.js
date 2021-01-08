const fs = require("fs");
const path = require("path");

const ordersServiceNamespaceYaml = fs.readFileSync(
  path.join(__dirname, "./ns.yaml"),
  {
    encoding: "utf8",
  }
);

const ordersServiceMicroserviceYaml = fs.readFileSync(
  path.join(__dirname, "./microservice.yaml"),
  {
    encoding: "utf8",
  }
);

const addonServiceBindingServiceInstanceYaml = fs.readFileSync(
  path.join(__dirname, "./redis-addon-sb-si.yaml"),
  {
    encoding: "utf8",
  }
);

const sbuYaml = fs.readFileSync(path.join(__dirname, "./redis-sbu.yaml"), {
  encoding: "utf8",
});

const serviceInstanceYaml = fs.readFileSync(
  path.join(__dirname, "./serviceinstance.yaml"),
  {
    encoding: "utf8",
  }
);

const xfMocksYaml = fs.readFileSync(path.join(__dirname, "./xf-mocks.yaml"), {
  encoding: "utf8",
});

module.exports = {
  sbuYaml,
  ordersServiceNamespaceYaml,
  ordersServiceMicroserviceYaml,
  serviceInstanceYaml,
  addonServiceBindingServiceInstanceYaml,
  xfMocksYaml,
};
