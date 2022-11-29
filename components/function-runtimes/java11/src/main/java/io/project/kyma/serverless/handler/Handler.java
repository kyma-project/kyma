package io.project.kyma.serverless.handler;

import javax.ws.rs.core.Context;
import javax.ws.rs.core.Response;

import io.project.kyma.serverless.sdk.CloudEventImpl;
import io.project.kyma.serverless.sdk.Function;


public class Handler implements Function {

    @Override
    public Response call(CloudEventImpl event, Context context) {
        throw new IllegalStateException("Not implemented stub Handler");
        //TODO: read data from request which will be passed to receiver
//        String receiverData = "";
//
//        String msgId = UUID.randomUUID().toString();
//        String eventType = "sap.kyma.custom.acme.payload.sanitised.v1";
//        String eventSource = "kyma";
//        Span span = null;
//        try {
//            span = event.tracer.spanBuilder("function triggered").startSpan();
//            var outEvent = event.buildResponseCloudEvent(msgId, eventType, receiverData);
//            outEvent.source = eventSource;
//            outEvent.specVersion = "1.0";
//            try {
//                event.publishCloudEvent(outEvent);
//                span.addEvent("Event sent");
//                span.setAttribute("event-type", eventType);
//                span.setAttribute("event-source", eventSource);
//                span.setAttribute("event-id", msgId);
//                span.setStatus(StatusCode.OK);
//            } catch (IOException | InterruptedException e) {
//                span.setStatus(StatusCode.ERROR, e.getMessage());
//                throw new RuntimeException(e);
//            }
//        } finally {
//            if (span != null) {
//                span.end();
//            }
//        }
//        return Response.ok(RETURN_STRING).build();
    }
}
