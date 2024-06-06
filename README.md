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

## Early thoughts on how to build a tool to analyze threats and attacks

TLDR;
- Do we aim to be a free exploration tool?
- How can we bring security speciallists knowledege?
- There are multiple type of relationships between Assets, how do we mesh them together? Do we need to mesh them? types:
  - Infrastructural State
  - Instrastructural Changes
  - Data communication
- Infrastructural State annalysis: it's dangerous. Networking and Authentication is complex and intricate. Do we
feel confident on stating "EC2 is not open to the internet"? What if we are wrong? I see this as deeper than a simple bug
- Stale Assets vs Deleted Assets. How to treat them in a "Time Series" like solution? An asset that didn't show up in the
previous iteration not necessarily means it's deleted (e.g. data source has an outage and we couldn't get that data). 


- Exploration is a very open term. Tinsae has shared this very nice tool called 
[Cartography](https://lyft.github.io/cartography/index.html) that essentially does what I tried to do (but better
and more complete). I find their approach interesting for us to consider. How they work is by fetching data from
a source and storing in Neo4j. From their, is up to the analyst to know how to write queries in Cypher (Neo4j 
query language).
  - From one side that requires the analyst to learn Cypher. From the other that empowers the user to explorer the data
  in the way they want.
  - What do we want to be? A free exploration tool like Neo4j? Or a guided exploration tool?
    - To be a free exploration tool inside elasticsearch would be a challenge. How to fit a free 
    exploration graph into ES? Is ES|QL enough? What if it's not? I fear for performance (gut feeling, we need to check)
    - If we are a free exploration tool, how do we make sure we covered every corner from the sources we have 
    in our graph? How do we keep up with data supplier latest developments?
    - I believe the previous point also applies to the guided exploration tool. But on the guided side it's easier to 
    think smaller. And claim that we support only `x,z,y` use cases. But still there is a need to make sure that 
    if we say "AWS Instance i-0000" is not accessible from the internet we considered VPCs, subnets, Route Tables, Load 
    Balancing, security groups, NAT Gateways, Network ACLs and others. There a lot of moving pieces.
    - If we are not a free exploration tool, we could have more freedom to model data in ES without thinking 
    "How is the easiest way for a user to query?". Or maybe that thought should always be there, regardless. 
- The knowledge to analyze the network structure in a Cloud will only partially translate to another Cloud Provider.
    It might actually be a challenge to have a simple code base wit such a goal
- The project on ES is really dependent on how ES implement joins. If it's fast I have good trust that modeling 
ES documents as nodes, and every node has a unidirectional, rich edge field will take us far.
- To partially automate the knowledge of a Security/Network specialist requires a specialist to code.
  I'm not a specialist.
  - As a non-specialist developer it will be fundamental that, first, I increase my knowledge, but second, we have
   a real specialist on the data source we are developing to consult during development. Ideally I, as the developer,
   eventually becomes a specialist. But I find this too critical to trust only on my knowledge. If we say "this machine 
   is not open", but in fact it's, that would be pretty bad for the project and elastic itself.
  - Even if we go towards the free exploration tool as a vision, we need to breakdown implementations per User Journey.
    my struggle with such a suggestion is: if we close User Journeys, can we confidently build up to be a free 
    exploration tool?
