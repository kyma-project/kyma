# When your certificates are created and replicated to kyma-system and you can access kyma console you can delete cert-manager

kubectl delete ns cert-manager
kubectl scale -n kyma-integration statefulset application-operator --replicas=0