```shell
docker build -t apc-api -f container/Dockerfile .
```

```shell
docker run -p 80:80 --name apc-api apc-api
```