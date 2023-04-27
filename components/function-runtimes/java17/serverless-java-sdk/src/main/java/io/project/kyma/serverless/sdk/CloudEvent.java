package io.project.kyma.serverless.sdk;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.ObjectWriter;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.propagation.TextMapSetter;
import jakarta.ws.rs.client.Client;
import jakarta.ws.rs.client.ClientBuilder;
import jakarta.ws.rs.client.Entity;
import jakarta.ws.rs.client.Invocation;
import jakarta.ws.rs.container.ContainerRequestContext;
import jakarta.ws.rs.core.MediaType;
import jakarta.ws.rs.core.MultivaluedHashMap;
import jakarta.ws.rs.core.MultivaluedMap;
import jakarta.ws.rs.core.Response;
import org.glassfish.jersey.client.ClientConfig;
import org.glassfish.jersey.logging.LoggingFeature;

import java.io.IOException;
import java.net.URI;
import java.util.Arrays;
import java.util.logging.Level;
import java.util.logging.Logger;

public class CloudEvent {

    private static final CloudEventHeaders[] CLOUD_EVENT_HEADERS = {CloudEventHeaders.CE_TYPE,
            CloudEventHeaders.CE_SOURCE,
            CloudEventHeaders.CE_EVENT_TYPE_VERSION,
            CloudEventHeaders.CE_SPEC_VERSION,
            CloudEventHeaders.CE_ID,
            CloudEventHeaders.CE_TIME,};

    private enum CloudEventHeaders {
        CE_TYPE("ce-type"), CE_SOURCE("ce-source"), CE_EVENT_TYPE_VERSION("ce-eventtypeversion"),
        CE_SPEC_VERSION("ce-specversion"), CE_TIME("ce-time"), CE_ID("ce-id");

        private final String headerName;

        public String getHeader() {
            return this.headerName;
        }

        CloudEventHeaders(String name) {
            this.headerName = name;
        }
    }

    public ContainerRequestContext req;
    public final MultivaluedMap<String, String> ceHeaders;

    public Tracer tracer;

    private final URI publishedProxyAddress;

    private final OpenTelemetry openTelemetry;

    public CloudEvent(ContainerRequestContext req, OpenTelemetry openTelemetry, Tracer tracer, URI publisherAddr) {
        this.req = req;
        this.tracer = tracer;
        this.ceHeaders = extractCloudEventHeaders(req.getHeaders());
        this.openTelemetry = openTelemetry;
        this.publishedProxyAddress = publisherAddr;
    }


    public ResponseCloudEvent buildResponseCloudEvent(String id, String type, String data) {
        var ceResponse = new ResponseCloudEvent();
        ceResponse.type = type;
        ceResponse.source = getHeaderValue(ceHeaders, CloudEventHeaders.CE_SOURCE);
        ceResponse.eventTypeVersion = getHeaderValue(ceHeaders, CloudEventHeaders.CE_EVENT_TYPE_VERSION);
        ceResponse.specVersion = getHeaderValue(ceHeaders, CloudEventHeaders.CE_SPEC_VERSION);
        ceResponse.id = id;
        ceResponse.data = data;
        ceResponse.dataContentType = resolveDataType(data);
        return ceResponse;
    }

    public void publishCloudEvent(ResponseCloudEvent ceEvent) throws IOException, InterruptedException {
        ObjectWriter ow = new ObjectMapper().writer();
        var outBody = ow.writeValueAsBytes(ceEvent.data);

        ClientConfig config = new ClientConfig();
        config.register(new LoggingFeature(Logger.getLogger(LoggingFeature.DEFAULT_LOGGER_NAME), Level.INFO, LoggingFeature.Verbosity.PAYLOAD_ANY, 10000));

        Client client = ClientBuilder.newClient(config);

        Invocation.Builder reqBuilder = client.target(this.publishedProxyAddress).request().
                header("Content-Type", "application/json").
                header(CloudEventHeaders.CE_SPEC_VERSION.getHeader(), ceEvent.specVersion).
                header(CloudEventHeaders.CE_TYPE.getHeader(), ceEvent.type).
                header(CloudEventHeaders.CE_SOURCE.getHeader(), ceEvent.source).
                header(CloudEventHeaders.CE_EVENT_TYPE_VERSION.getHeader(), ceEvent.eventTypeVersion).
                header(CloudEventHeaders.CE_ID.getHeader(), ceEvent.id);

        injectHeaderSetter(reqBuilder);
        var res = reqBuilder.post(Entity.json(ceEvent.data));
        if (Response.Status.Family.familyOf(res.getStatus()) != Response.Status.Family.SUCCESSFUL) {
            throw new IOException("Failed to send event. The publisher responded with:" + res.getStatus() + "status code which is not in 2xx successful family");
        }
    }

    public Invocation.Builder getTraceableRequestBuilder(String target) {
        Client client = ClientBuilder.newClient();
        Invocation.Builder reqBuilder = client.target(target).request();
        injectHeaderSetter(reqBuilder);
        return reqBuilder;
    }

    private void injectHeaderSetter(Invocation.Builder reqBuilder) {

        TextMapSetter<Invocation.Builder> setter = (carrier, key, value) -> {
            // Insert the context as Header
            System.out.println("Inject->" + key + ":" + value);
            assert carrier != null;
            carrier.header(key, value);
        };
        openTelemetry.getPropagators().getTextMapPropagator().inject(io.opentelemetry.context.Context.current(), reqBuilder, setter);
    }

    private static MultivaluedMap<String, String> extractCloudEventHeaders(MultivaluedMap<String, String> headers) {
        MultivaluedMap<String, String> ceHeaders = new MultivaluedHashMap<>();
        Arrays.stream(CLOUD_EVENT_HEADERS).forEach(ceHeader -> ceHeaders.add(ceHeader.getHeader(), getHeaderValue(headers, ceHeader)));
        return ceHeaders;
    }

    private static String getHeaderValue(MultivaluedMap<String, String> headers, CloudEventHeaders ceHeader) {
        String headerValue = "";
        var headerValues = headers.get(ceHeader.getHeader());
        if (headerValues != null && headerValues.size() > 0) {
            headerValue = headerValues.get(0);
        }
        return headerValue;
    }


    private static MediaType resolveDataType(String data) {
        try {
            final ObjectMapper mapper = new ObjectMapper();
            mapper.readTree(data);
            return MediaType.APPLICATION_JSON_TYPE;

        } catch (IOException ignored) {

        }
        return MediaType.TEXT_PLAIN_TYPE;
    }

}
