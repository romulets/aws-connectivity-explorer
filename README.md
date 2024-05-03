[![Go](https://github.com/romulets/aws-connectivity-explorer/actions/workflows/go.yml/badge.svg)](https://github.com/romulets/aws-connectivity-explorer/actions/workflows/go.yml)

# AWS Connectivity Explorer
Experimental project to learn about AWS Networking and Graph Databases

## Use cases

- Store all instances in a region and correlate by VPC id `POST /ec2-instances/fetch-graph`
- Fetch all instances with public IP and SSH port open `GET /ec2-instances/ssh-open-to-internet`
- Fetch all instances in the same VPC as another instance `GET /ec2-instances/in-vpc/{instanceId}`

## How to run

1. [Configure AWS credentials locally](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-quickstart.html). 
For ease of development, the default local aws credentials configured in the machine are being used
2. Run Neo4J
```shell
docker run \
    --publish=7474:7474 --publish=7687:7687 \
    --volume=$HOME/neo4j/data:/data \
    neo4j
```
3. Configure `config.yml`. You can copy [config.example.yml](config.example.yml)
4. Run the project
```
    go run main.go
```

## Code Structure

- `application/` holds everything related to serve HTTP requests
- `core/` contains fetching, grouping and storage of aws assets. Why storage, you might ask. It felt
like the core logic of this app also lives inside the database. I can be wrong of course. But for simplicity
I followed my gut feeling. Regardless `aws` package is completely decoupled from anything else in this project.
That is the real core.
- `support/` things that support the application to run, like configuration, concurrency management among others