1. Generate certificate and key using this script: https://github.com/alex-leonhardt/k8s-mutate-webhook/blob/master/ssl/ssl.sh
2. Create secret in kyma-integration namespace containing generated certificate and key.
3. Build image of Application Operator
4. Edit Application Operator by replacing image and adding volume mount pointing on the secret:
   
   ```
   volumes:
     - name: webhook-cert
       secret:
         defaultMode: 420
         items:
         - key: tls.crt
           path: webhook.crt
         - key: tls.key
           path: webhook.key
         secretName: SECRET_NAME
   ``` 
5. Create service from file resources/service.yaml
6. Populate caBundle in file resources/webhook.yaml. You can retrieve caBundle using this command: 
```kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}'```
7. Create MutatingWebhookConfiguration from file resources/webhook.yaml
8. Create Application and bind it to namespace. Delete Application, Application Mapping should be also deleted.