const { expect } = require("chai");
const https = require("https");

const config = require('./config.js');

describe("Function", () => {
    it("should return Hello World", (done) => {
        https.get(config.functionUrl, (res) => {
            expect(res.statusCode).to.be.equal(200);
            res.on('data', (data) => {
                expect(data.toString()).to.be.equal('Hello World!');
                done();
            });
        });
    });
});
