FROM node:14.16.1-alpine3.12

LABEL source = git@github.com:kyma-project/kyma.git

COPY kubeless/ /

WORKDIR /kubeless_rt/

RUN npm install

USER 1000

CMD ["node", "kubeless.js"]