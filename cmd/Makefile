test:
	GOOS=linux go build -o clustering .
	docker build -t clustering:dev .
	docker run -d --name node-1 -e CLUSTER_PORT=10731 -e CLUSTER_SEED_NODES=172.17.0.2 clustering:dev
	docker run -d --name node-2 -e CLUSTER_PORT=10731 -e CLUSTER_SEED_NODES=172.17.0.2 clustering:dev
	docker run -d --name node-3 -e CLUSTER_PORT=10731 -e CLUSTER_SEED_NODES=172.17.0.2 clustering:dev
	docker run -d --name node-4 -e CLUSTER_PORT=10731 -e CLUSTER_SEED_NODES=172.17.0.2 clustering:dev
	docker run -d --name node-5 -e CLUSTER_PORT=10731 -e CLUSTER_SEED_NODES=172.17.0.2 clustering:dev
	docker run -d --name node-6 -e CLUSTER_PORT=10731 -e CLUSTER_SEED_NODES=172.17.0.2 clustering:dev
	docker run -d --name node-7 -e CLUSTER_PORT=10731 -e CLUSTER_SEED_NODES=172.17.0.2 clustering:dev
	docker run -d --name node-8 -e CLUSTER_PORT=10731 -e CLUSTER_SEED_NODES=172.17.0.2 clustering:dev
	docker run -d --name node-9 -e CLUSTER_PORT=10731 -e CLUSTER_SEED_NODES=172.17.0.2 clustering:dev
clean:
	docker stop node-1
	docker stop node-2
	docker stop node-3
	docker stop node-4
	docker stop node-5
	docker stop node-6
	docker stop node-7
	docker stop node-8
	docker stop node-9
	docker system prune -f