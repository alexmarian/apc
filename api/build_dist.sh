mkdir disttmp
mkdir disttmp/static
cd ../ui && npm install && npm run build && cd ../api
cp -R ../ui/dist/* disttmp/static/
docker build -t goapp-export --target export-stage -f dist/Dockerfile .
docker create --name transfer_container goapp-export /bin/bash
docker cp transfer_container:/main ./disttmp/main
docker rm transfer_container
zip -r dist.zip disttmp/*
rm -rf disttmp
