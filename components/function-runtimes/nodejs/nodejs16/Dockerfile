# image base on 16.19.1-alpine3.17
FROM node@sha256:43b162893518666b4a08d95dae49153f22a5dba85c229f8b0b8113b609000bc2

ARG NODE_ENV
ENV NODE_ENV $NODE_ENV
ENV npm_config_cache /tmp/

RUN mkdir -p /usr/src/app
RUN mkdir -p /usr/src/app/lib
WORKDIR /usr/src/app

COPY ./nodejs16/package.json /usr/src/app/
RUN npm install && npm cache clean --force
COPY ../lib /usr/src/app/lib

COPY ../server.js /usr/src/app/server.js

CMD ["npm", "start"]

EXPOSE 8888
