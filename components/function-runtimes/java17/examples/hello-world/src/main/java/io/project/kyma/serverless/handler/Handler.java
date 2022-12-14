package io.project.kyma.serverless.handler;

import javax.ws.rs.core.Context;
import javax.ws.rs.core.Response;

import io.project.kyma.serverless.sdk.CloudEvent;
import io.project.kyma.serverless.sdk.Function;


public class Handler implements Function {

    public static final String RETURN_STRING = "Hello World from java17 runtime with serverless SDK!";

    @Override
    public Response main(CloudEvent event, Context context) {
        return Response.ok(RETURN_STRING).build();
    }
}
