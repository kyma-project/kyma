import logging
from locust import HttpUser, task, constant, events

expected_string = "Hello Serverless"

class Python38Test(HttpUser):
    wait_time = constant(1)

    @task
    def hello_word(self):
        with self.client.get("/data", catch_response=True) as response:
            if expected_string not in response.text:
                response.failure("Got wrong response, expected {} ..., got: {}".format(expected_string,response.text))
