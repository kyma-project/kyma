const chai = require("chai");
const chaiHttp = require("chai-http");
const { expect } = chai;
chai.use(chaiHttp);

const config = require('./config.js');

describe("Function", () => {
    it("should return Hello World", (done) => {
        chai.request(config.functionUrl)
        .get('/')
        .end((_, res) => {
            expect(res.status).to.be.equal(200);
            expect(res.text).to.be.equal('Hello World!');
            done();
        });
    });
});
