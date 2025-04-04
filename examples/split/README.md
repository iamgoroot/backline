# Backline Split Example

This example shows how to: 
- collect entities during docker image build into sqlite database
- index collected entities into local files (using bluge plugin)
- deploy webapp bundled with collected during build entities and indices
- protect webapp with oauth2 authentication

Example use cases (or adaptations):

1. Your organization has strict security policies that don't allow you to use any external resources on production environment.
2. You want to trigger entity scanning only after certain event (e.g. CI build)
3. You want to deploy webapp separately but run entity scanning separately from webapp


### Scan service

This service downloads entities from github and stores them in sqlite database. It also indexes them into bluge index saving index into a directory.

```bash
go run ./examples/split/scan_service/main.go --config ./examples/split/scan_service/config.yaml
```


### Webapp

Compile webapp binary and put it with sqlite database and bluge index into docker image.

```bash
CGO_ENABLED=0 GOOS=linux go build -o backline ./examples/split/webapp/main.go
```

Then webapp can be run with config file that points to sqlite database and bluge index.

This example is used to deploy demo version of Backline