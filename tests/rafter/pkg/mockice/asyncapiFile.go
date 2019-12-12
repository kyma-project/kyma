package mockice

const AsyncAPIFile = `
asyncapi: '1.0.0'
info:
  title: Streetlights API
  version: '1.0.0'
  description: |
    The Smartylighting Streetlights API allows you to remotely manage the city lights.

    ### Check out its awesome features:

    * Turn a specific streetlight on/off 🌃
    * Dim a specific streetlight 😎
    * Receive real-time information about environmental lighting conditions 📈
  license:
    name: Apache 2.0
    url: https://www.apache.org/licenses/LICENSE-2.0
baseTopic: smartylighting.streetlights.1.0

servers:
  - url: api.streetlights.smartylighting.com:{port}
    scheme: mqtt
    description: Test broker
    variables:
      port:
        description: Secure connection (TLS) is available through port 8883.
        default: '1883'
        enum:
          - '1883'
          - '8883'

security:
  - apiKey: []

topics:
  event.{streetlightId}.lighting.measured:
    parameters:
      - $ref: '#/components/parameters/streetlightId'
    publish:
      $ref: '#/components/messages/lightMeasured'

  action.{streetlightId}.turn.on:
    parameters:
      - $ref: '#/components/parameters/streetlightId'
    subscribe:
      $ref: '#/components/messages/turnOnOff'

  action.{streetlightId}.turn.off:
    parameters:
      - $ref: '#/components/parameters/streetlightId'
    subscribe:
      $ref: '#/components/messages/turnOnOff'

  action.{streetlightId}.dim:
    parameters:
      - $ref: '#/components/parameters/streetlightId'
    subscribe:
      $ref: '#/components/messages/dimLight'

components:
  messages:
    lightMeasured:
      summary: Inform about environmental lighting conditions for a particular streetlight.
      payload:
        $ref: "#/components/schemas/lightMeasuredPayload"
    turnOnOff:
      summary: Command a particular streetlight to turn the lights on or off.
      payload:
        $ref: "#/components/schemas/turnOnOffPayload"
    dimLight:
      summary: Command a particular streetlight to dim the lights.
      payload:
        $ref: "#/components/schemas/dimLightPayload"

  schemas:
    lightMeasuredPayload:
      type: object
      properties:
        lumens:
          type: integer
          minimum: 0
          description: Light intensity measured in lumens.
        sentAt:
          $ref: "#/components/schemas/sentAt"
    turnOnOffPayload:
      type: object
      properties:
        command:
          type: string
          enum:
            - on
            - off
          description: Whether to turn on or off the light.
        sentAt:
          $ref: "#/components/schemas/sentAt"
    dimLightPayload:
      type: object
      properties:
        percentage:
          type: integer
          description: Percentage to which the light should be dimmed to.
          minimum: 0
          maximum: 100
        sentAt:
          $ref: "#/components/schemas/sentAt"
    sentAt:
      type: string
      format: date-time
      description: Date and time when the message was sent.

  securitySchemes:
    apiKey:
      type: apiKey
      in: user
      description: Provide your API key as the user and leave the password empty.

  parameters:
    streetlightId:
      name: streetlightId
      description: The ID of the streetlight.
      schema:
        type: string
`
