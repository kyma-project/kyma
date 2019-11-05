# Alerting rules

## Overview

Kyma uses Prometheus alerting rules to monitor the health of resources and inform about potential issues.

## Details

Kyma comes with a set of alerting rules provided out of the box. You can find them [here](https://github.com/kyma-project/kyma/tree/master/resources/monitoring/charts/alertmanager/templates). These rules provide alerting configuration for logging, webapps, rest services, and custom Kyma rules. You can also create your own alerting rules.

### Create alerting rules

Follow [this](https://kyma-project.io/docs/components/monitoring/#tutorials-define-alerting-rules) tutorial to create alerting rules.


### Configure Alertmanager

Configure the Alertmanager using the [alertmanager](../alertmanager/README.md) chart.
