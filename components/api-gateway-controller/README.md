# Api-Gateway Controller (name to be changed)


## Example CRD Overview:

```yaml
---
application:
  service:
    name: foo-service
    port: 8080
  hostURL: https://foo.bar
authentication:
  type: 
  - name: JWT
    config:
      issuer: http://dex.kyma.local
      jwks: []
      mode: 
      - name: ALL
        config:
          scopes: []
      - name: EXCLUDE
        config:
          - pathSuffix: '/c'
          - pathRegex: '/d/*'
          - pathPrefix: ''
          - pathExact: '/f/foobar.png'
      - name: INCLUDE
        config:
          - path: '/a'
            scopes: 
              - read
            methods:
              - GET
              - POST
          - path: '/b'
            methods:
              - GET
  - name: PASSTHROUGH
    config: {}  
  - name: OAUTH
    config:
      - path: '/a'
        scopes: 
          - write
        methods:
          - POST
      # Invalid or takes priority
      - path: '/*' 
        scopes: []
        methods: []

```