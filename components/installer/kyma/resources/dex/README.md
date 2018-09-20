```
  _  __                       _____            
 | |/ /                      |  __ \           
 | ' /_   _ _ __ ___   __ _  | |  | | _____  __
 |  <| | | | '_ ` _ \ / _` | | |  | |/ _ \ \/ /
 | . \ |_| | | | | | | (_| | | |__| |  __/>  <
 |_|\_\__, |_| |_| |_|\__,_| |_____/ \___/_/\_\
       __/ |                                   
      |___/                                               
```

## Overview

Kyma Dex is an identity service that delegates user authentication to external identity providers using [connectors](https://github.com/coreos/dex#connectors).
For more details about Dex, see the [Dex GitHub](https://github.com/coreos/dex) project.

## Details

Currently, Kyma Dex uses a static user database and authenticates static users by itself, instead of using a fully-integrated authentication solution. Kyma Dex also comes with a static list of clients allowed to initiate the OAuth2 flow.

For the list of static Dex users and clients, as well as the information about the connectors that delegate authentication to external identity providers, see the [dex-config-map.yaml](templates/dex-config-map.yaml) file.

Kyma Dex is exposed using the [Istio VirtualService](https://istio.io/docs/reference/config/istio.networking.v1alpha3/#VirtualService) feature. Access Dex at  `https://dex.kyma.local`.
