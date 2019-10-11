# Grab Tech Talks - Go Concurrency

## Components

- zookeeper
- kafka-broker
- prometheus
- grafana
- locust
- payment service

```
$ kafka-console-consumer --bootstrap-server localhost:29092 --topic payment
$ GOOS=linux GOARCH=amd64 go build -o payment-service .
$ docker build -t payment-service:0.5.0 .
$ docker run --rm -d --name payment-service -p 8100:8100 -p 9100:9100 --memory="1024m" --cpus="2" --network host payment-service:0.5.0
$ docker run --rm -it --name payment-service -p 8100:8100 -p 9100:9100 --memory="1024m" --cpus="2" --network host payment-service:0.5.0 bash
$ docker run --rm -it --name payment-service --memory="1024m" --cpus="2" --network host payment-service:0.5.0 bash
```

## Guidelines

```sh
docker-compose up -d

docker run -d -p 8000:8000 -p 9000:9000 --name portainer --network bridge -v /var/run/docker.sock:/var/run/docker.sock -v portainer_data:/data portainer/portainer
docker run -d -p 9090:9090 --name prometheus --network bridge -v ~/.bin/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus
docker run -d -p 3000:3000 --name grafana --network bridge grafana/grafana

docker run -d --rm -p 9200:9200 --name payment-service --network bridge --memory="1024m" --cpus="2" payment-service:0.5.1
docker run -d --rm -p 9200:9200 --name payment-service --network bridge --memory="1024m" --cpus="2" payment-service:0.5.2
docker run -d --rm -p 9200:9200 --name payment-service --network bridge --memory="1024m" --cpus="2" payment-service:0.5.3

docker run -it --rm -p 9200:9200 --name payment-service --network bridge --memory="1024m" --cpus="2" payment-service:0.5.0 bash
```

