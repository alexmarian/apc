cd ../ui && npm install && npm run build && cd ../api
cp -R ../ui/dist/ static/
docker build -t apc-api -f container/Dockerfile .
