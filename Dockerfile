FROM golang:1.20
WORKDIR /app

# install dependencies
RUN apt-get update && apt-get install -y git npm
RUN apt-get install -y build-essential
RUN apt-get install -y libseccomp-dev

# install ddp
ARG ddppath
ENV DDPPATH=/app/DDP
ENV PATH=/app/DDP:${PATH}
ADD ${ddppath} /app/DDP

# clone the repo
RUN git clone https://github.com/DDP-Projekt/Spielplatz.git
WORKDIR /app/Spielplatz
RUN npm install
COPY config.json ./

# run the app
EXPOSE 8001
CMD make && ./run.sh