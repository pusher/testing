ECHO := echo -e

all: verify-config

.PHONY: config
config: verify-config run-config verify-config

.PHONY: run-config
run-config:
	@ $(ECHO) "\033[36mGenerating Config\033[0m"
	kubectl create configmap config --from-file=config.yaml=config/config.yaml -n default --dry-run -o yaml > prow/config.yaml
	kubectl create configmap plugins --from-file=plugins.yaml=config/plugins.yaml -n default --dry-run -o yaml > prow/plugins.yaml
	scripts/make-jobs-config.sh > prow/jobs.yaml
	@for f in config plugins jobs; do printf '#############\n###\n### THIS IS AN AUTOGENERATED FILE!!! DO NOT EDIT THIS FILE DIRECTLY!!!\n###\n#############\n\n%s\n' "$$(cat prow/$${f}.yaml)" > prow/$${f}.yaml; done
	@ echo # Produce a new line at the end of each target to help readability

.PHONY: verify-config
verify-config:
	@ $(ECHO) "\033[36mVerifying Config\033[0m"
	docker run --rm -v $(shell pwd)/config:/config gcr.io/k8s-prow/checkconfig:v20200220-18fae0a00 --config-path=/config/config.yaml --job-config-path=/config/jobs --plugin-config=/config/plugins.yaml
	@ echo # Spacer between output
	make run-config
	@ $(ECHO) "\033[36mVerifying Git Status\033[0m"
	@ if [ "$$(git status -s)" != "" ]; then git diff --color; $(ECHO) "\033[31;1mERROR: Git Diff found. Please run \`make config\` and commit the result.\033[0m"; exit 1; else $(ECHO) "\033[32mValid config found\033[0m";fi
	@ echo # Produce a new line at the end of each target to help readability

.PHONY:
check-image-tags:
	@ $(ECHO) "\033[36m\033[1mChecking image tags\033[0m"
	scripts/check-image-tags.sh
	@ echo # Produce a new line at the end of each target to help readability

TAG ?= v20190821-328974b
.PHONY:
update-image-tags:
	@ $(ECHO) "\033[36m\033[1mUpdating image tags\033[0m"
	scripts/update-image-tags.sh $(TAG)
	@ echo # Produce a new line at the end of each target to help readability


.PHONY:
verify-image-tags: update-image-tags check-image-tags
	@ $(ECHO) "\033[36m\033[1mVerifying Git Status\033[0m"
	@ if [ "$$(git status -s)" != "" ]; then git diff --color; $(ECHO) "\033[31m\033[1mERROR: Git Diff found. Please run \`make update-image-tags\` and commit the result.\033[0m"; exit 1; else $(ECHO) "\033[32mAll image tags verified\033[0m";fi
	@ echo # Produce a new line at the end of each target to help readability
