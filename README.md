# Poker Server with javascript client

Poker engine by me but much of the cards deck rules come from https://github.com/notnil/joker

***

**THERE ARE A TON OF BUGS!!!  But might provide someone with inspiration, good luck.

**Javscript side is largely a hack job

## build
    make build

## Build Image

Requires AWS ECR to push image.  Create your ECR repository and update the account and repo details in Makefile

    make build-image
   
## Run

Running will need AWS credentials to connect to dynamodb to write table info and hands

    docker run -it -p 8080:8080 -v ~/.aws:/.aws poker-service:latest