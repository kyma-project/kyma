package io.project.kyma.serverless;

import io.opentelemetry.api.GlobalOpenTelemetry;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.internal.StringUtils;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.SpanKind;
import io.opentelemetry.context.propagation.TextMapGetter;
import io.project.kyma.serverless.handler.Handler;
import io.project.kyma.serverless.sdk.CloudEvent;
import io.project.kyma.serverless.sdk.Function;

import jakarta.ws.rs.*;
import jakarta.ws.rs.container.ContainerRequestContext;
import jakarta.ws.rs.core.Context;
import jakarta.ws.rs.core.MultivaluedMap;
import jakarta.ws.rs.core.Response;
import java.net.URI;
import java.util.logging.Logger;

@Path("/")
public class JerseyServer {


    private final Function fn;

    private static final Logger logger = Logger.getGlobal();

    private final URI publisherProxyAddr;

    private final OpenTelemetry openTelemetry;
    private final String svcName;

    public JerseyServer(OpenTelemetry openTelemetry, URI publisherProxyAddr, String svcName) {
        this.publisherProxyAddr = publisherProxyAddr;
        this.svcName = svcName;
        this.openTelemetry = openTelemetry;
        this.fn = new Handler();
    }

    @GET
    @Path("/healthz")
    public Response healthz(@Context ContainerRequestContext request) {
        return Response.ok("ok").build();
    }

    @GET
    public Response home(@Context ContainerRequestContext request) {
        return callUserFunction(request);
    }

    @POST
    public Response homePost(@Context ContainerRequestContext request) {
        return callUserFunction(request);
    }

    @PUT
    public Response homePut(@Context ContainerRequestContext request) {
        return callUserFunction(request);
    }

    @DELETE
    public Response homeDelete(@Context ContainerRequestContext request) {
        return callUserFunction(request);
    }


    private Response callUserFunction(ContainerRequestContext httpRequest) {
        var tracer = openTelemetry.getTracerProvider().get(svcName);
        var extractedContext = injectPropagatorGetter(httpRequest);
        extractedContext.makeCurrent();
        Span span = null;
        try {
            span = tracer.spanBuilder("request").setSpanKind(SpanKind.SERVER).startSpan();
            span.makeCurrent();

            var ceEvent = new CloudEvent(httpRequest, openTelemetry, tracer, this.publisherProxyAddr);
            return this.fn.main(ceEvent, null);
        } finally {
            if (span != null) {
                span.end();
            }
        }
    }

    private io.opentelemetry.context.Context injectPropagatorGetter(ContainerRequestContext httpRequest) {
        TextMapGetter<ContainerRequestContext> getter = new TextMapGetter<>() {
            @Override
            public Iterable<String> keys(ContainerRequestContext requestContext) {
                return requestContext.getHeaders().keySet();
            }

            @Override
            public String get(ContainerRequestContext requestContext, String key) {
                String value = getHeaderValue(requestContext.getHeaders(), key);
                if (StringUtils.isNullOrEmpty(value)) {
                    return null;
                }
                return value;
            }
        };
        return GlobalOpenTelemetry.get().getPropagators().getTextMapPropagator().extract(io.opentelemetry.context.Context.current(), httpRequest, getter);
    }

    private static String getHeaderValue(MultivaluedMap<String, String> headers, String key) {
        String headerValue = "";
        var headerValues = headers.get(key);
        if (headerValues != null && headerValues.size() > 0) {
            headerValue = headerValues.get(0);
        }
        return headerValue;
    }
}

