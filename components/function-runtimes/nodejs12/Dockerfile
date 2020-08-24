FROM kubeless/nodejs@sha256:aa819e9bee6c22db0b137595ac45765ef1cad64d0dc09502a008ce4446c607ef as source

FROM node:12.18.3-alpine3.12

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=source /kubeless_rt/ /kubeless_rt/
COPY kubeless-npm-install.sh /

WORKDIR /kubeless_rt/

RUN npm install

USER 1000

CMD ["node", "kubeless.js"]