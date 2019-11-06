FROM golang:1.12.10-alpine3.10 as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/components/application-broker
ADD . $SRC_DIR

WORKDIR $SRC_DIR

RUN CGO_ENABLED=0 go build -o /application-broker ./cmd/broker

RUN wget

FROM scratch


#RUN apk --no-cache add ca-certificates
#RUN apk add --no-cache curl

# Variables used for labeling created images
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder ./application-broker /root/application-broker
COPY ./licenses/ /root/licenses

ENTRYPOINT ["/root/application-broker"]
