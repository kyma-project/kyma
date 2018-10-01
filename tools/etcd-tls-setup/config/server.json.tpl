{
    "CN": "etcd server",
    "hosts": [
        "*.__ETCD_CLUSTER_NAME__",
        "*.__ETCD_CLUSTER_NAME__.__NAMESPACE__",
        "*.__ETCD_CLUSTER_NAME__.__NAMESPACE__.svc",
        "*.__ETCD_CLUSTER_NAME__.__NAMESPACE__.svc.cluster.local",
        "__ETCD_CLUSTER_NAME__-client.__NAMESPACE__.svc.cluster.local",
        "localhost"
    ],
    "key": {
        "algo": "rsa",
        "size": 2048
    },
    "names": [
        {
            "C": "PL",
            "L": "SL",
            "ST": "Kyma"
        }
    ]
}
