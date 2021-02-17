# E2E Subscriber

This is a sample subscriber app that simulates an event subscriber via exposing a set of APIs which can be used to receive and list events. It is meant to be only used in tests.


## Endpoints

| Endpoint                                     | Method | Usage                                                                                             |
|-------------------------------------------- |------ |------------------------------------------------------------------------------------------------- |
| `/ce`                                      | POST   | Stores the received CloudEvent.                                                              |
| `/ce`                                      | GET    | Returns a list of all received CloudEvents.                                                           |
| `/ce/by-uuid/<uuid>`                               | GET    | Returns a CloudEvent with a matching UUID. Returns `200` on success, `204` if no event was found.        |
| `/ce/<source>/<type>/<event-type-version>` | GET    | Returns all CloudEvents matching the source/type/version. Returns `200` on success, `204` if no event was found. |
| `/`                                        | POST   | Increases the counter of received requests. CloudEvents sent to `/ce` do not increase this counter.     |
| `/`                                        | GET    | Returns the current counter value.                                                                     |
| `/`                                        | DELETE | Resets the counter and clears the list of received CloudEvents.                                                    |

## Commandline

| Flag                   | Usage          |
|---------------------- |-------------- |
| `--port <int>` | Listens on a provided port.|
