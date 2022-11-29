package io.project.kyma.serverless.sdk;

import javax.ws.rs.client.Invocation;

public interface TraceableRequest {
    Invocation.Builder getTraceableRequestBuilder(String target);

}