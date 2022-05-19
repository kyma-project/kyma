# Performance with multiple pipelines

Here we tried to perform perfomance tests to understand the impact on memory, cpu and throughput by increasing the number of pipelines.

We performed tests with following number of `pipelines`: 0,1,2,3,5,10,20. In the test we would measure the `memory/cpu` consumption immediately after deploying the pipelines and then we would measure these values again after 10 mins. For generating huge amount of logs we used a `log spammer` and we used a `http logging backend` (mockserver) in a different kyma cluster.

## The setup

Following is the setup

![a](./assets/setup.drawio.svg)



The setup of fluentbit was following

![a](./assets/setup-3c.drawio.svg)


## Summary

Below is the summary of the tests performed for more detailed results check the graphs section below

1. Idle CPU is .31
2. Increasing CPU increases the  rate at which inputs are read
3. The avg CPU usage is around 1.2 for upto 20 pipelines. The CPU plateaus after 2 pipelines
4. Memory is usually not a problem for low number of pipelines. Had to increase memory to 512 MB for 10 pipelines
5. The memory increases linearly
6. With 1.5CPU/512 Mi memory we can support 20 pipelines (under extreme conditions)
7. With two pipelines, when one pipeline (http server) is down we see that the output and emitter throughput goes down to 0. However  checking the filesystem buffer the chunks are rolled. There are no logs saying why the output stops functioning.


## Learnings
1. [Health check](https://docs.fluentbit.io/manual/administration/monitoring#health-check-for-fluent-bit) had to be adjusted as due to higher error rate the fluentbit would mark the pod unhealthy.
2. To improve the [performance](https://www.mock-server.com/mock_server/performance.html) of mockserver we had to recude the memory consumption the logging the messages to stdout was disabled. Also the log level was changed to `trace`.
3. Over period of time the http output throguhput was declining, however after restarting mockserver it was increasing again. It looks like a issue with mockserver.
4. Loki was running out of memory had to increase the memory of Loki to 1Gi.
5. Increasing the [http worker count](https://docs.fluentbit.io/manual/pipeline/outputs/http) to 10 increased the CPU usage but did not have any affect on the throughput.


## Results

1. CPU cores utilization

    ![a](./assets/cpu-cores.jpg)


2. Memory utilization

    ![a](./assets/memory.jpg)

3. Input throughput all pipelines are good

    ![a](./assets/input-throughput-100-good.jpg)

4. Output throughput 50% pipelines are bad
    
    ![a](./assets/input-throughput-50-bad.jpg)

5. Output throughput all pipelines are good

    ![a](./assets/output-throughput-100-good.jpg)

6. Output throughput 50% pipelines are bad

    ![a](./assets/output-throughput-50-bad.jpg)