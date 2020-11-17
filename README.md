## News Parser
This application allows you to parse and aggregate news from multiple websites according to passed patterns.

### Getting started
With docker-compose: `make`
Without docker:
- install postgresql
- install [migrating tool](https://github.com/golang-migrate/migrate)
- run ```migrate -database 'postgresql://mylogin:mypassword@localhost:5432/mydbname?sslmode=disable' -path ./migrations up```
with your credentials
- run ```go build -o ./parser && ./parser -dbURL 'postgresql://mylogin:mypassword@localhost:5432/mydbname?sslmode=disable'```

### Endpoints
See [parser.proto](https://github.com/IlyushaZ/parser/blob/master/api/proto/parser.proto)

### TODO:
- Add unit tests