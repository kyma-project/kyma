Local installation explained in details

This document assumes that reader checked-out kyma repository. All scripts are fired and discussed in context of root path. 

To fire local installation run following command:
```./installation/cmd/run.sh```

This script will set up default parameters, start minikube, build kyma-installer, generate local config, create installer custom resource and finally set up installer

This script can be fired with following flags:

--skip-minikube-start