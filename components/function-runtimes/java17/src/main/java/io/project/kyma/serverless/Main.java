package io.project.kyma.serverless;

import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.context.propagation.ContextPropagators;
import io.opentelemetry.context.propagation.TextMapPropagator;
import io.opentelemetry.exporter.otlp.http.trace.OtlpHttpSpanExporter;
import io.opentelemetry.extension.trace.propagation.B3Propagator;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.sdk.trace.SdkTracerProvider;
import io.opentelemetry.sdk.trace.export.SimpleSpanProcessor;
import io.opentelemetry.semconv.resource.attributes.ResourceAttributes;
import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.servlet.ServletContextHandler;
import org.eclipse.jetty.servlet.ServletHolder;
import org.glassfish.jersey.server.ResourceConfig;
import org.glassfish.jersey.servlet.ServletContainer;

import java.net.URI;
import java.util.Arrays;
import java.util.stream.Collectors;

public class Main {

    public Main(Config config) throws Exception {
        String svcName = createSvcName(config.podName, config.serviceNamespace);

        var openTelemetry = configureTracing(config.tracingCollectorAddr, svcName);
        Server server = configureServer(config.port, openTelemetry, svcName, config.publisherProxyAddr);
        server.start();
        server.join();
    }

    private Server configureServer(int serverPort, OpenTelemetry openTelemetry, String svcName, URI publisherProxyAddr) {
        ResourceConfig resourceConfig = new ResourceConfig();

        JerseyServer jerseyServer = new JerseyServer(openTelemetry, publisherProxyAddr, svcName);
        resourceConfig.registerInstances(jerseyServer);

        ServletContainer servletContainer = new ServletContainer(resourceConfig);
        ServletHolder sh = new ServletHolder(servletContainer);

        ServletContextHandler context = new ServletContextHandler(ServletContextHandler.SESSIONS);
        context.addServlet(sh, "/*");

        Server server = new Server(serverPort);
        server.setHandler(context);
        return server;
    }

    static String createSvcName(String podName, String svcNamespace) {
        if ((podName == null) || (svcNamespace == null)) {
            return "generic-svc";
        }
        // remove generated pods suffix ( two last sections )
        var svcNameBuilder = Arrays.stream(podName.split("-")).limit(2).
                collect(Collectors.joining("-"));
        return String.join(".", svcNameBuilder, svcNamespace);
    }

    private OpenTelemetry configureTracing(URI tracingEndpoint, String svcName) {
        Resource resource = Resource.getDefault()
                .merge(Resource.create(Attributes.of(ResourceAttributes.SERVICE_NAME, svcName)));

        SdkTracerProvider sdkTracerProvider = SdkTracerProvider.builder()
                .addSpanProcessor(SimpleSpanProcessor.create(OtlpHttpSpanExporter.builder().setEndpoint(tracingEndpoint.toString()).build()))
                .setResource(resource)
                .build();
        TextMapPropagator b3Propagator = B3Propagator.injectingMultiHeaders();
        var sdk = OpenTelemetrySdk.builder().setPropagators(ContextPropagators.create(b3Propagator)).
                setTracerProvider(sdkTracerProvider).
                buildAndRegisterGlobal();
        return sdk;
    }

    public static void main(String[] args) throws Exception {
        Config config = new Config();
        new Main(config);
    }
}
