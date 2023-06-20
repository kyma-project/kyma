# image base on 16.20.0-alpine3.17.4
FROM node@sha256:9811fa24a30dd363b0de9cc67467970c5aefbd84615d4981b833526735903e05

ARG NODE_ENV
ENV NODE_ENV $NODE_ENV
ENV npm_config_cache /tmp/

RUN mkdir -p /usr/src/app
RUN mkdir -p /usr/src/app/lib
WORKDIR /usr/src/app

COPY ./nodejs16/package.json /usr/src/app/
RUN npm install && npm cache clean --force
COPY ./lib /usr/src/app/lib

COPY ./server.js /usr/src/app/server.js

CMD ["npm", "start"]

EXPOSE 8888
