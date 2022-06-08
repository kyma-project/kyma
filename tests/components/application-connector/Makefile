IMAGE = ""

.PHONY: image publish

test-gateway: publish
	helm template deployments/gateway-test --set testImage=$(IMAGE) | kubectl apply -f -
	kubectl wait --for=condition=complete job/myjob &
	completion_pid=$$!
	kubectl wait --for=condition=failed job/myjob && exit 1 &
	failure_pid=$$!
	wait $$completion_pid $$failure_pid

clean:
	helm template deployments/gateway-test --set testImage=$(IMAGE) | kubectl delete -f -

publish: image
	docker push $(IMAGE)

image:
	docker build . -t $(IMAGE)
