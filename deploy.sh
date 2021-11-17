echo "Deploying..."

git pull
make down
make build
make up

echo "Deployment completed"
