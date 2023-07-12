# image base on 16.20.1-alpine3.17.4
FROM node@sha256:b1fdeade9cae98c30bbd8087f26f8da404e6fc48bdd53772017855c1a1d32605


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
