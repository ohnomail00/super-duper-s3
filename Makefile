export PROJECT="super-duper-s3"

imports-sort:
	@echo "Reformatting imports..."
	@go install github.com/daixiang0/gci@latest
	@gci write --skip-generated --skip-vendor  -s standard -s default -s 'prefix(github.com/ohnomail00/super-duper-s3)' -s localmodule .

docker-build:
	@echo "Building docker image..."
	@docker build --ssh default \
		--target builder \
		-t ${PROJECT}-gateway:build \
		-f build/gateway.Dockerfile .
	@docker build --ssh default \
		--target builder \
		-t ${PROJECT}-storage:build \
		-f build/storage.Dockerfile .

docker-build-if-nex:
ifeq ($(HAS_BUILD_IMAGE), 0)
	$(MAKE) docker-build
else
	@echo "Build image exists"
endif

docker-run: docker-build-if-nex
	@echo "Starting server in docker container..."
	@docker-compose -f build/docker_compose.yml up -d

docker-down:
	@docker-compose -f build/docker_compose.yml down

docker-network:
	@docker network create "${PROJECT}" || true

docker-network-down:
	@docker network rm "${PROJECT}" || true

docker-infra: docker-network
	@docker-compose -f build/docker_compose_infra.yml up -d

docker-infra-down:
	@docker-compose -f build/docker_compose_infra.yml stop
	@$(MAKE) docker-network-down
