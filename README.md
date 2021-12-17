
# iudx-sandbox-backend

## Get Started

### Prerequisite - Make configuration

Create a `.env` file in the root directory of the project based on the template given below:

```sh
# Server config
API_PORT=8080

# DB config
POSTGRES_USER=user
POSTGRES_PASSWORD=user123
POSTGRES_HOST=pg
POSTGRES_PORT=5432
POSTGRES_DB=sandbox_db
```

### Docker based

1. Install docker and docker-compose.
2. Clone this repo.
3. Build the images
   `./docker/build.sh`
4. Modify the `docker-compose.yml` file to map the config file you just created.
5. Start the server in production (prod) or development (dev) mode using docker-compose
   ` docker-compose up prod `

To bring down all the services, run `docker-compose down`

### Testing

### Unit tests

Run the tests through docker using docker-compose
   ` docker-compose up test `
