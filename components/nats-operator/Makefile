##############
#            #
#   common   #
#            #
##############

SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles
include $(SCRIPTS_DIR)/generic-make-go.mk
TAG = $(DOCKER_TAG)

.PHONY: release
release:
	$(MAKE) gomod-release-local

.PHONY: build-image
build-image: build-image-nats-operator build-image-nats-operator-doctor ;

.PHONY: push-image
push-image: push-image-nats-operator push-image-nats-operator-doctor ;

#####################
#                   #
#   nats-operator   #
#                   #
#####################

NATS_OPERATOR_APP_NAME = nats-operator
NATS_OPERATOR_IMG_NAME = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(NATS_OPERATOR_APP_NAME)

.PHONY: build-image-nats-operator
build-image-nats-operator:
	docker build -t $(NATS_OPERATOR_APP_NAME):latest -f cmd/nats-operator/Dockerfile .

.PHONY: push-image-nats-operator
push-image-nats-operator:
	docker tag $(NATS_OPERATOR_APP_NAME) $(NATS_OPERATOR_IMG_NAME):$(TAG)
	docker push $(NATS_OPERATOR_IMG_NAME):$(TAG)

############################
#                          #
#   nats-operator-doctor   #
#                          #
############################

NATS_OPERATOR_DOCTOR_APP_NAME = nats-operator-doctor
NATS_OPERATOR_DOCTOR_IMG_NAME = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(NATS_OPERATOR_DOCTOR_APP_NAME)

.PHONY: build-image-nats-operator-doctor
build-image-nats-operator-doctor: pull-licenses
	docker build -t $(NATS_OPERATOR_DOCTOR_APP_NAME):latest -f cmd/nats-operator-doctor/Dockerfile .

.PHONY: push-image-nats-operator-doctor
push-image-nats-operator-doctor:
	docker tag $(NATS_OPERATOR_DOCTOR_APP_NAME) $(NATS_OPERATOR_DOCTOR_IMG_NAME):$(TAG)
	docker push $(NATS_OPERATOR_DOCTOR_IMG_NAME):$(TAG)
