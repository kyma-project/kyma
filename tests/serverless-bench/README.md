# Serverless-Bench

The Serverless-Bench image is based on the `locustio/locust` image. It also includes a locust test configuration and a simple script to run the test and output the results in JSON to stdout. The logs are collected via a log sink and are pushed to Google Cloud BigQuery for further analysis.  