## Alem Core

## 1. Introduction

This go backend project dedicated to make REST API for mobile app called "Alem", that helps to find jobs in Europe for CIS employeers. The techonlogy used there is go-chi as the REST API framework, postgresql as the main database.

#### Techonlogy

- Backend: Go, chi
- Database: Postgresql
- Authentication: JWT

#### Api Base Url

https://dev.jumystap.kz/api/v1/docs

## 2. Getting Started

#### Prerequisites

- Install go
- Install dependecies

#### Running the project

```
    make run
```

#### Envirement variables

- config/local.yaml

```
    env: "local"
    port: 8000
    database:
        user: "postgres"
        password: "password"
        host: "localhost"
        port: 5432
        sslmode: "disable"
        name: "alem"
```

## 3. Project Structure

```
/alem
│── /cmd                 # Entry point of the project
│   ├── /alem            # Entry point of the REST API
│   ├── /migrate         # Entry point of the migration logic
│
│── /internal            # Internal logic files that will not be imported
│   ├── /app             # All routes of the project
│   ├── /http            # Logic for handling HTTP requests
│   │   ├── /handler     # HTTP handlers
│   │   ├── /middleware  # HTTP middleware
│   ├── /service         # Business logic of the app
│   ├── /repository      # Data handling layer
│   ├── /model           # Structs for interacting with the database
│   ├── /dto             # Structs for HTTP interactions
│
│── /migrations          # Migration files
```

#### TODO

- [X] Resume Experience repository layer
- [X] Resume service layer with CRUD
- [X] Resume handler
- [X] Organization services migration
- [ ] Organization services model/dto
- [ ] Organization services repository CRUD
- [ ] Organization services service layer CRUD
- [ ] Organization services handler
- [X] Write documentation
- [ ] Make the swagger
