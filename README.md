# restd
RESTful API and web server daemon

Building in docker
==================

MUSL target
-----------

```
docker-compose -f build/docker-compose.build.yml up --build musl-local
```

glibc target
-----------

```
docker-compose -f build/docker-compose.build.yml up --build glibc-local
```

