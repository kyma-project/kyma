# image base on 16.18.1-alpine3.16
FROM node@sha256:34d5b3bc90f9beb47e7eb14c17837c9d393eb2524b34c9421b1d675ba05b69b2

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
