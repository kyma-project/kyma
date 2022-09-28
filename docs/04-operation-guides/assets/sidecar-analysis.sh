if [ "$1" == "" ]; then
    echo "Pods out of istio mesh:"

    echo "  In namespace labeled with \"istio-injection=disabled\":"
    disabled_namespaces=$(kubectl get ns -l "istio-injection=disabled" -o jsonpath='{.items[*].metadata.name}')

    for ns in $disabled_namespaces
    do
        if [ "$ns" != "kube-system" ] && [ "$ns" != "kyma-system" ]; then
            pods_out_of_istio=$(kubectl get pod -n $ns -o jsonpath='{.items[*].metadata.name}')
            for pod in $pods_out_of_istio
            do
                echo "    - $ns/$pod"
            done
        fi
    done

    echo "  In namespace labeled with \"istio-injection=enabled\" with pod labeled with \"sidecar.istio.io/inject=false\":"
    enabled_ns=$(kubectl get ns -l "istio-injection=enabled" -o jsonpath='{.items[*].metadata.name}')
    for ns in $enabled_ns
    do
        if [ "$ns" != "kube-system" ] && [ "$ns" != "kyma-system" ]; then
            pods_out_of_istio=$(kubectl get pod -l "sidecar.istio.io/inject=false" -n $ns -o jsonpath='{.items[*].metadata.name}')
            for pod in $pods_out_of_istio
            do
                echo "    - $ns/$pod"
            done
        fi
    done

    echo "  In not labeled ns with pod not labeled with \"sidecar.istio.io/inject=true\":"
    not_labeled_ns=$(kubectl get ns -l "istio-injection!=disabled, istio-injection!=enabled" -o jsonpath='{.items[*].metadata.name}')
    for ns in $not_labeled_ns
    do
        if [ "$ns" != "kube-system" ] && [ "$ns" != "kyma-system" ]; then
            pods_out_of_istio=$(kubectl get pod -l "sidecar.istio.io/inject!=true" -n $ns -o jsonpath='{.items[*].metadata.name}')
            for pod in $pods_out_of_istio
            do
                echo "    - $ns/$pod"
            done
        fi
    done
else 
    ns_label=$(kubectl get ns $1 -o jsonpath='{.metadata.labels.istio-injection}')
    echo "Pods out of istio mesh in namespace $1:"
    if [ "$ns_label" == "enabled" ]; then
        pods_out_of_istio=$(kubectl get pod -l "sidecar.istio.io/inject=false" -n $1 -o jsonpath='{.items[*].metadata.name}')
        for pod in $pods_out_of_istio
        do
            echo "  - $pod"
        done
    elif [ "$ns_label" == "disabled" ]; then
        pods_out_of_istio=$(kubectl get pod -n $1 -o jsonpath='{.items[*].metadata.name}')
        for pod in $pods_out_of_istio
        do
            echo "  - $pod"
        done
    else
        pods_out_of_istio=$(kubectl get pod -l "sidecar.istio.io/inject!=true" -n $1 -o jsonpath='{.items[*].metadata.name}')
        for pod in $pods_out_of_istio
        do
            echo "  - $pod"
        done
    fi
fi