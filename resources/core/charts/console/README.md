```
   _____                      _      
  / ____|                    | |     
 | |     ___  _ __  ___  ___ | | ___
 | |    / _ \| '_ \/ __|/ _ \| |/ _ \
 | |___| (_) | | | \__ \ (_) | |  __/
  \_____\___/|_| |_|___/\___/|_|\___|

```
## Overview

The Console is a web-based UI for Kyma.
It allows users to manage specific functionality within Kyma along with basic Kubernetes resources. 
The Console provides an extensibility mechanism which allows you to seamlessly integrate UI parts to achieve additional functionality.

## Details

This section provides details related to the configuration of the Console.

### Configuration

The deployment of the Console includes a [ConfigMap](templates/configmap.yaml). 
The ConfigMap introduces a `config.js` file that is mounted as the asset of the Console application and injected as a configuration file. 
Use this mechanism to overwrite the default configuration with custom values resulting from the Helm chart installation.
