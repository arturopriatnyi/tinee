echo "Deploying..."

git pull
docker-compose down
docker-compose up --build

echo "Deployment completed"
