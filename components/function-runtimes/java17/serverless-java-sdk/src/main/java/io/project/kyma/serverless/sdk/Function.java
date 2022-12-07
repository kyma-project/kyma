package io.project.kyma.serverless.sdk;

import javax.ws.rs.core.Context;
import javax.ws.rs.core.Response;


public interface Function {
    Response call(CloudEvent event, Context context);
}
