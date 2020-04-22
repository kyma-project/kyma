FROM kubeless/nodejs@sha256:aa819e9bee6c22db0b137595ac45765ef1cad64d0dc09502a008ce4446c607ef as source

FROM node:12.16.2-alpine3.11

COPY --from=source /kubeless_rt/ /kubeless_rt/
COPY --from=source /kubeless-npm-install.sh /

WORKDIR /kubeless_rt/

RUN npm install

USER 1000

CMD ["node", "kubeless.js"]