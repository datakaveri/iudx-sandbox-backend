
# iudx-sandbox-backend

## Get Started

### Prerequisite - Make configuration

Create a `.env` file in the root directory of the project

### Docker based

1. Install docker and docker-compose.
2. Clone this repo.
3. Build the images
   `./docker/build.sh`
4. Modify the `docker-compose.yml` file to map the config file you just created.
5. Start the server in production (prod) or development (dev) mode using docker-compose
   ` docker-compose up prod `

### Testing

### Unit tests

Run the tests through docker using docker-compose
   ` docker-compose up test `
