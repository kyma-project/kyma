package io.project.kyma.serverless.handler;

import javax.ws.rs.core.Context;
import javax.ws.rs.core.Response;

import io.project.kyma.serverless.sdk.CloudEvent;
import io.project.kyma.serverless.sdk.Function;


public class Handler implements Function {

    @Override
    public Response main(CloudEvent event, Context context) {
        throw new IllegalStateException("Not implemented stub Handler");
    }
}
