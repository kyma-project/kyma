# E2E Subscriber

This is a sample subscriber app that simulates an event subscriber via exposing a set of APIs which can be used to receive and list events. It is meant to be only used in tests.


## Endpoints

| Endpoint                                     | Method | Usage                                                                                             |
|-------------------------------------------- |------ |------------------------------------------------------------------------------------------------- |
| `/ce`                                      | POST   | stores the received cloudevent                                                                    |
| `/ce`                                      | GET    | return list of all received cloudevents                                                           |
| `/ce/<uuid>`                               | GET    | return cloud event with matching uuid. Returns 200 on success, 204 if no event was found          |
| `/ce/<source>/<type>/<event-type-version>` | GET    | return all events matching source/type/version. Returns 200 on success, 204 if no event was found |
| `/`                                        | POST   | increase counter of received requests. Cloudevents sent to /ce do not increase this counter       |
| `/`                                        | GET    | return current counter value                                                                      |
| `/`                                        | DELETE | reset counter and list of received cloudevents                                                    |

## Commandline

| Flag                   | Usage          |
|---------------------- |-------------- |
| `--port <int>` | listen on port |
