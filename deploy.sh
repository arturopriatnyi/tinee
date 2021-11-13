echo "Deploying..."

git pull
docker-compose down
docker-compose up -d --build

echo "Deployment completed"
