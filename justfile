# migrate-db creates an sql file in ./migrations directory where migration queries should be written.
migrate-db FILENAME:
  goose -dir ./migrations create {{FILENAME}} sql

start:
	docker compose up -d

stop:
	docker compose down

build:
	docker compose build

rebuild: stop build start

docker-deploy:
    docker build -t nutri-price-server .
    docker tag nutri-price-server eu.gcr.io/nutriprice/nutri-price-server:latest
    docker push eu.gcr.io/nutriprice/nutri-price-server:latest

deploy:
    gcloud run deploy nutri-price-server \
    --image eu.gcr.io/nutriprice/nutri-price-server \
    --set-secrets DATABASE_URL=DATABASE_URL:latest \
    --platform managed \
    --region europe-west1 \
    --allow-unauthenticated

build-and-deploy: docker-deploy deploy

test:
	go test ./...