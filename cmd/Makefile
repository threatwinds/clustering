test:
	GOOS=linux go build -o clustering .
	docker build -t clustering:dev .
	zsh -c 'for i in {1..30}; do docker run -d --name node-$$i -e CLUSTER_PORT=10731 -e CLUSTER_SEED_NODES=172.17.0.2,172.17.0.3,172.17.0.4,172.17.0.5 clustering:dev; done'

clean:
	zsh -c 'for i in {1..30}; do docker stop node-$$i; done'
	docker system prune -f