import http from 'k6/http';
import {Trend} from "k6/metrics";
import {check, group} from "k6";

export let options = {
    vus: `${__ENV.VUS}`,
    duration: "2m",
    tags: {
        "testName": `istio_load_test-${__ENV.VUS}-vus`,
        "component": "istio",
        "revision": `${__ENV.REVISION}`
    },
    conf: {
        workerCount: `${__ENV.WORKLOAD_SIZE}`,
        domain: `${__ENV.CLUSTER_DOMAIN_NAME}`,
        namespace: `${__ENV.NAMESPACE}`
    }
}

const manyToManyName = `vus-${options.vus}-many-to-many`;
let manyToMany = new Trend(manyToManyName, true);

const manyToSingleName = `vus-${options.vus}-one-to-httpbin-1`;
let manyToSingle = new Trend(manyToSingleName, true);

const manyToOwnName = `vus-${options.vus}-many-to-own`;
let manyToOwn = new Trend(manyToOwnName, true);

export default function () {
    group("Each VUS Calls all Services", function () {
        const workers = 1 + parseInt(options.conf.workerCount);
        for (let i = 1; i < workers; i++) {
            const response = http.get(`https://httpbin-${i}.${options.conf.namespace}.${options.conf.domain}/cookies`);
            manyToMany.add(response.timings.duration);

            check(response, {
                "status was 200": (r) => r.status === 200,
                "status was NOT 501": (r) => r.status !== 501,
                "status was NOT 503": (r) => r.status !== 503,
                "transaction time < 150 ms": (r) => r.timings.duration < 150,
                "transaction time in (150;400) ms": (r) => (r.timings.duration > 150 && r.timings.duration < 400),
                "transaction time > 400 ms": (r) => r.timings.duration > 400,
            }, {
                "test": manyToManyName,
            });
        }
    });

    group("Each VUS Calls httpbin-1 Service", function () {
        const response = http.get(`https://httpbin-1.${options.conf.namespace}.${options.conf.domain}/cookies`);
        manyToSingle.add(response.timings.duration);

        check(response, {
            "status was 200": (r) => r.status === 200,
            "status was NOT 501": (r) => r.status !== 501,
            "status was NOT 503": (r) => r.status !== 503,
            "transaction time < 150 ms": (r) => r.timings.duration < 150,
            "transaction time in (150;400) ms": (r) => (r.timings.duration > 150 && r.timings.duration < 400),
            "transaction time > 400 ms": (r) => r.timings.duration > 400,
        }, {
            "test": manyToSingleName,
        });
    });

    group("Each VUS Calls own Service", function () {
        const vu = parseInt(`${__VU}`);
        const bins = parseInt(options.conf.workerCount);
        const bin = (vu % bins) + 1;

        const response = http.get(`https://httpbin-${bin}.${options.conf.namespace}.${options.conf.domain}/cookies`);
        manyToOwn.add(response.timings.duration);

        check(response, {
            "status was 200": (r) => r.status === 200,
            "status was NOT 501": (r) => r.status !== 501,
            "status was NOT 503": (r) => r.status !== 503,
            "transaction time < 150 ms": (r) => r.timings.duration < 150,
            "transaction time in (150;400) ms": (r) => (r.timings.duration > 150 && r.timings.duration < 400),
            "transaction time > 400 ms": (r) => r.timings.duration > 400,
        }, {
            "test": manyToOwnName,
        });
    });
}
