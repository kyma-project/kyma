package runtimes

import (
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
)

const basicJavaRequirements = `
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">

    <modelVersion>4.0.0</modelVersion>
    <groupId>io.project.kyma.serverless</groupId>
    <artifactId>hello-world</artifactId>
    <version>0.0.1</version>
    <packaging>jar</packaging>

    <name>hello-world</name>

    <properties>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <maven.compiler.source>11</maven.compiler.source>
        <maven.compiler.target>11</maven.compiler.target>
    </properties>

    <dependencies>
        <dependency>
            <groupId>javax.ws.rs</groupId>
            <artifactId>javax.ws.rs-api</artifactId>
            <version>2.0</version>
            <scope>provided</scope>
        </dependency>
        <dependency>
            <groupId>io.project.kyma.serverless</groupId>
            <artifactId>serverless-java-sdk</artifactId>
            <version>0.0.1</version>
            <scope>compile</scope>
        </dependency>
    </dependencies>
</project>
`

const functionCodeTpl = `
package io.project.kyma.serverless.handler;
        
        import javax.ws.rs.core.Context;
        import javax.ws.rs.core.Response;
        import io.project.kyma.serverless.sdk.CloudEvent;
        import io.project.kyma.serverless.sdk.Function;
        
        
        public class Handler implements Function {
        
            public static final String RETURN_STRING = "Hello World from local java11 runtime from docker graalvm with serverless SDK!";
        
            @Override
            public Response call(CloudEvent event, Context context) {
                return Response.ok("%s").build();
            }
        }
`

func BasicJavaFunction(returnMsg string, rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	var spec = serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       fmt.Sprintf(functionCodeTpl, returnMsg),
				Dependencies: basicJavaRequirements,
			},
		},
	}

	return spec
}

func BasicJavaFunctionWithCustomDependency(returnMsg string, rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	var spec = serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source: fmt.Sprintf(functionCodeTpl, returnMsg),
				//TODO: add basic java dependency
				Dependencies: basicJavaRequirements,
			},
		},
	}

	return spec
}
