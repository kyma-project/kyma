package io.project.kyma.serverless;

import static org.junit.jupiter.api.Assertions.*;

class MainTest {

    @org.junit.jupiter.api.Test
    void createSvcName_Success() {
        //GIVEN
        String svcName = "default";
        String podName = "emitter-qqmds-84dd76fc94-2pnpd";
        String expected = "emitter-qqmds.default";
        //WHEN
        String output = Main.createSvcName(podName, svcName);
        //THEN
        assertEquals(expected, output);
    }
}