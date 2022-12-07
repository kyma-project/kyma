package io.project.kyma.serverless;

import java.net.URI;
import java.net.URISyntaxException;

public class Config {

    private static final int DEFAULT_PORT = 8080;
    protected final URI publisherProxyAddr;
    protected final URI tracingCollectorAddr;
    protected int port;
    protected final String podName;
    protected final String serviceNamespace;

    protected Config() throws IllegalArgumentException {
        this.publisherProxyAddr = getURIFromEnv("PUBLISHER_PROXY_ADDRESS");
        this.tracingCollectorAddr = getURIFromEnv("TRACE_COLLECTOR_ENDPOINT");
        this.podName = System.getenv("HOSTNAME");
        this.serviceNamespace = System.getenv("SERVICE_NAMESPACE");
        this.port = getNumber("FUNCTION_PORT");
    }

    private int getNumber(String envName) {
        int serverPort = DEFAULT_PORT;
        String fnPort = System.getenv(envName);
        if (fnPort != null && fnPort.equals("")) {
            serverPort = Integer.parseInt(fnPort);
        }
        return serverPort;
    }

    private URI getURIFromEnv(String envName) throws IllegalArgumentException {
        String envValue = System.getenv(envName);
        if (envValue == null) {
            throw new IllegalArgumentException("Couldn't find env:" + envName);
        }
        try {
            return new URI(envValue);
        } catch (URISyntaxException e) {
            throw new IllegalArgumentException("Couldn't parse env:" + envName + "with value:" + envValue, e);
        }
    }
}
