package io.project.kyma.serverless.handler;

import javax.ws.rs.core.Context;
import javax.ws.rs.core.Response;
import io.project.kyma.serverless.sdk.CloudEventImpl;
import io.project.kyma.serverless.sdk.Function;


public class Handler implements Function {

    public static final String RETURN_STRING = "Hello World from local java11 runtime from docker graalvm with serverless SDK!";

    @Override
    public Response call(CloudEventImpl event, Context context) {
        return Response.ok(RETURN_STRING).build();
    }
}
