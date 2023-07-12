# image base on node:18.16.1-alpine3.17 (node:18.16.0, alpine:3.17.4)
FROM node@sha256:56a82eb7c721b7e4c151366000a60d1f1b1ded51e55e8ff9cc106f603dfe6521


ARG NODE_ENV
ENV NODE_ENV $NODE_ENV
ENV npm_config_cache /tmp/

RUN mkdir -p /usr/src/app
RUN mkdir -p /usr/src/app/lib
WORKDIR /usr/src/app

COPY ./nodejs18/package.json /usr/src/app/
RUN npm install && npm cache clean --force
COPY ./lib /usr/src/app/lib

COPY ./server.js /usr/src/app/server.js

CMD ["npm", "start"]

EXPOSE 8888
