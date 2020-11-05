# NATS Chart

This Helm chart deploy NATS: https://nats.io/
Steps:
- Create "nats" namespace:
```bash
$ kubectl create ns  install nats
```
- Install it using Helm 3 on "nats" namespace:
```bash
$ helm install nats nats -n nats --set global.install.enabled=true
```
- Test the installation:
```bash
$ kubectl -n nats port-forward nats-1 4222
```
Open two terminals.
In the first one, create "foo" subject and subscribe to it using client ID 1:
```bash
$ telnet localhost 4222
SUB foo 1
```
In the second terminal, publish a message to the "foo" subject. The lenght of the message must be snet before the content
```bash
$ telnet localhost 4222
PUB foo 10
Hello Kyma
PUB foo 3
Bye
```
- Unintall nats:
```bash
helm uninstall nats -n nats
```

