FROM golang:bookworm as build
WORKDIR /

# install dependencies
RUN apt-get update && apt-get install -y \
	npm \
	build-essential \
	libseccomp-dev

ARG DDP_VERSION

ADD https://github.com/DDP-Projekt/Kompilierer/releases/latest/download/DDP-$DDP_VERSION-linux-amd64.tar.gz ./DDP.tar.gz
RUN mkdir /DDP
RUN tar -xzf ./DDP.tar.gz -C DDP --strip-components 1

ENV DDPPATH=/DDP
ENV CGO_ENABLED=1

COPY ./ /
RUN make DDPVERSION=${DDP_VERSION}

FROM debian:bookworm-slim as run

WORKDIR /

RUN apt-get update && apt-get install -y \
	libseccomp-dev \
	gcc \
	libtinfo-dev \
	libpcre2-dev \
	locales

RUN echo "de_DE.UTF-8 UTF-8" > /etc/locale.gen
RUN echo "LANG=de_DE.UTF-8" > /etc/default/locale
RUN locale-gen de_DE.UTF-8

COPY --from=build /DDP /DDP
WORKDIR /DDP
RUN chmod +x ./ddp-setup
RUN ./ddp-setup --force

ENV PATH=/DDP/bin:$PATH
ENV DDPPATH=/DDP

COPY --from=build Spielplatz seccomp_exec seccomp_main.o /app/
COPY --from=build /node_modules /app/node_modules
COPY --from=build /static/ /app/static

RUN rm -rf /var/cache/apt/archives /var/lib/apt/lists/*

WORKDIR /app

ENV GIN_MODE=release
EXPOSE 8001
CMD [ "/app/Spielplatz" ]