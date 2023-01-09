package io.project.kyma.serverless.sdk;

import jakarta.ws.rs.core.Context;
import jakarta.ws.rs.core.Response;


public interface Function {
    Response main(CloudEvent event, Context context);
}
