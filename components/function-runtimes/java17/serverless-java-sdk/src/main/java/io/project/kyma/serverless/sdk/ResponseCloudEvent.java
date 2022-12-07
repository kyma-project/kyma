package io.project.kyma.serverless.sdk;
import javax.ws.rs.core.MediaType;

public class ResponseCloudEvent {
    public String type;
    public String source;
    public String eventTypeVersion;
    public String specVersion;
    public String id;
    public String data;
    public MediaType dataContentType;

}
