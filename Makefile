DATE := $(shell date "+%Y%m%d%H%M")
RANDOM := $(shell bash -c 'echo $$RANDOM')
TEMPLATE_DIR := _output/template
REGISTRY ?= ccr.ccs.tencentyun.com/qcloud-ti-platform
VERSION  ?= v0.1-$(DATE)-$(RANDOM)
IMAGE    ?= $(REGISTRY)/device-plugin-demo:$(VERSION)
BUILD_ARGS := --platform=linux/amd64 -t $(IMAGE)
HELM_ARGS := -n default --set image=$(IMAGE)
DEPLOY_YAML := $(TEMPLATE_DIR)/deploy.yaml

.DEFAULT_GOAL := all
.PHONY: all container template

all: clean container template

container:
	docker build $(BUILD_ARGS) .
	docker push $(IMAGE)

template: $(DEPLOY_YAML)

$(DEPLOY_YAML): deploy/device-plugin-demo
	mkdir -p $(TEMPLATE_DIR)
	helm template $(HELM_ARGS) device-plugin-demo $< > $@

deploy:
	k apply -f _output/template/deploy.yaml

clean:
	rm -rf $(TEMPLATE_DIR)