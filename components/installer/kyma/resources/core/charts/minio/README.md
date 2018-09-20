```
  __  __ _       _       
 |  \/  (_)     (_)      
 | \  / |_ _ __  _  ___  
 | |\/| | | '_ \| |/ _ \
 | |  | | | | | | | (_) |
 |_|  |_|_|_| |_|_|\___/


```

## Overview

Minio is an open source object storage server with Amazon S3 compatible API. Kyma provides Minio as a core component to store static content. For example, documentation, images, or videos. The size of an object can range from a few KBs to a maximum of 5TB. In the long term, you can replace Minio with an external solution, such as AWS S3.  

## Details

This section describes how to use Minio. Learn how to connect to Minio through a web browser or Minio Client.

### Connect to Minio through a web browser

From your browser, go to https://minio.kyma.local and log in with these credentials:
 - name: admin
 - password: topSecretKey

### Connect to Minio through Minio Client

1. Install Minio Client:
```
sudo apt-get install wget
wget https://dl.minio.io/client/mc/release/linux-amd64/mc
chmod a+x mc
```

2. Configure Minio Client:
```
./mc config host add myminio https://minio.kyma.local admin topSecretKey
```

3. Try out Minio Client:
```
./mc mb myminio/bucket1
ls
./mc cp mc myminio/bucket1
./mc ls myminio
./mc ls myminio/bucket1
```
