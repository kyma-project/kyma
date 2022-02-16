FROM debian

RUN apt update && apt install wget -y

COPY fetch.sh /app/
ENTRYPOINT /app/fetch.sh
