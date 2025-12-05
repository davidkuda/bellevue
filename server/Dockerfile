FROM golang:tip-trixie

RUN mkdir /app
WORKDIR /app
RUN echo ${PWD} && ls -lR


# npm packages
RUN apt update; apt install npm -y
RUN npm install; npm install esbuild

COPY package.json /app/
COPY Makefile /app/
COPY make/ /app/make/
COPY make/docker_bundle.mk /app/make/bundle.mk
COPY ui /app/ui

RUN make bundle/css; make bundle/js

COPY go.mod /app/
COPY internal /app/internal
COPY /cmd/web /app/cmd/web
COPY main.go /app/

RUN go mod tidy

CMD [ "go", "run", "./cmd/web" ]