tar czf ./DDP.tar.gz -C $DDPPATH .
docker build --build-arg ddppath=./DDP.tar.gz  --tag ddp-spielplatz .
