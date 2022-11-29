package io.project.kyma.serverless.sdk;

import java.io.IOException;

public interface CloudEvent {
    ResponseCloudEvent buildResponseCloudEvent(String id, String type, String data);

    void publishCloudEvent(ResponseCloudEvent ceEvent) throws IOException, InterruptedException;
}
