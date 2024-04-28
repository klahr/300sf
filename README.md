# 300 Svenska Fåglar (300 Swedish Birds)
This is an API designed for bird watching that manages the recording of bird observations associated with user accounts.

## Development
This guide provides an introduction to developing the project.

### Getting The Source-code
```bash
git clone https://git.sr.ht/~klahr/300sf
```

### Building
```bash
cd 300sf
go build -o bin/300sf
```

### Setting Up The Database
...

### Prepare Running
Copy the `config.yml.sample` file, renaming it to `config.yml`. Subsequently, update the `USERNAME` and `PASSWORD` fields within this file to correspond with your database credentials.
```bash
cp config.yml.sample config.yml
```

### Running
```
./bin/300sf -config ./config.yml
```

### Testing
```bash
curl localhost:8090/health
```
#### Response If Working
```json
 {"message":"healthy","data":{"version":"0.1.0"}}
 ```