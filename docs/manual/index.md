# Manual

## Setup Database

Install `mariadb` and `CREATE DATABASE 300sf`.

## Configure Application

### Generate a Public/Private Key Pair

TODO

### Create Config File

Create a new file in the root directory and name it `config.yml`. Populate it with the following:

```yml
---
data-source: ...
```

## Run Application

```bash
go build
./300sf -config ./config.yml
```

### (Optional) Test Calling The Server

```bash
curl localhost:8090/ping
```
