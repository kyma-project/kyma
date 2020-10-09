FROM node:12.18.4-alpine3.12

LABEL source = git@github.com:kyma-project/kyma.git

COPY kubeless/ /

WORKDIR /kubeless_rt/

RUN npm install

USER 1000

CMD ["node", "kubeless.js"]