# image base on 16.19.0-alpine3.17
FROM node@sha256:e427ffd1ba7915ca8d8aeba45806ef3f1f1e6b65ce152b363645cd7428f8d75a

ARG NODE_ENV
ENV NODE_ENV $NODE_ENV
ENV npm_config_cache /tmp/

RUN mkdir -p /usr/src/app
RUN mkdir -p /usr/src/app/lib
WORKDIR /usr/src/app

COPY package.json /usr/src/app/
RUN npm install && npm cache clean --force
COPY lib /usr/src/app/lib

COPY server.js /usr/src/app/server.js

CMD ["npm", "start"]

EXPOSE 8888
