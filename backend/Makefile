gen:
	goctl api go -api ./api/hackathon.api -dir . -remote git@github.com:Z3Labs/go-zero-template.git

format:
	goctl api format -dir ./api

run:
	go run hackathon.go -f ./etc/hackathon-api.yaml

all:
	make gen && make format

docker:
	chmod +x ./internal/clients/diagnosis/py/build-docker.sh && ./internal/clients/diagnosis/py/build-docker.sh

model:
	@if [ -z "$(type)" ]; then \
		echo "Error: type parameter is required"; \
		echo "Usage: make model type=<model_name>"; \
		exit 1; \
	fi
	go-gen model mongo --type $(type) --dir ./internal/model --template ../../go-zero-template/template