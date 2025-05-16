```shell
docker build -t apc-api -f container/Dockerfile .
```

```shell
docker run -p 80:8080 --name apc-api apc-api
```