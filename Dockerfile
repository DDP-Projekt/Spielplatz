FROM golang:1.22.2
WORKDIR /

# install dependencies
RUN apt-get update && apt-get install -y git \
	npm \
	build-essential \
	libseccomp-dev \
	libpcre2-dev \
	locales \
	libncurses5 \
	libz-dev \
	libtinfo-dev \
	libxml2-dev

RUN echo "de_DE.UTF-8 UTF-8" > /etc/locale.gen
RUN echo "LANG=de_DE.UTF-8" > /etc/default/locale
RUN locale-gen de_DE.UTF-8

# install llvm
WORKDIR /llvm
ARG llvm_archive
COPY ${llvm_archive} ./
RUN tar -xvf *.tar.* -C ./ --strip-components 1
RUN rm *.tar.*
ENV PATH=/llvm/bin:${PATH}

# clone the Kompilierer repo
WORKDIR /
RUN git clone https://github.com/DDP-Projekt/Kompilierer.git --depth=1
ENV DDPPATH=/Kompilierer/build/DDP
ENV PATH=/Kompilierer/build/DDP/bin:${PATH}

# install ddp
WORKDIR /Kompilierer
RUN go mod tidy
RUN make LLVM_CONFIG=llvm-config

# clone the repo
WORKDIR /app
RUN git clone https://github.com/DDP-Projekt/Spielplatz.git --depth=1
WORKDIR /app/Spielplatz
RUN npm install
RUN go mod tidy

# clone the config
COPY config.json ./
ARG certpath
ARG keypath
COPY ${certpath} ./
COPY ${keypath} ./

# configure git to use https instead of ssh
RUN git config --global url."https://github.com/".insteadOf git@github.com:
RUN git config --global url."https://".insteadOf git://

# run the app
ENV GIN_MODE=release
EXPOSE 8001
CMD  cd /Kompilierer && \
	git pull origin master && \
	go mod tidy && \
	make LLVM_CONFIG=llvm-config && \
	cd /app/Spielplatz && \
	git pull origin main && \
	go mod tidy && \
	./run.sh