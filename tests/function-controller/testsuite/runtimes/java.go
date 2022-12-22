package runtimes

import (
	"fmt"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
)

const basicJavaDepsTpl = `
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
            <groupId>jakarta.ws.rs</groupId>
            <artifactId>jakarta.ws.rs-api</artifactId>
            <version>3.1.0</version>
            <scope>provided</scope>
        </dependency>
        <dependency>
            <groupId>io.project.kyma.serverless</groupId>
            <artifactId>serverless-java-sdk</artifactId>
            <version>0.0.1</version>
            <scope>compile</scope>
        </dependency>
		%s
    </dependencies>
</project>
`

const additionalLib = `
<!-- https://mvnrepository.com/artifact/com.fasterxml.jackson.dataformat/jackson-dataformat-csv -->
<dependency>
    <groupId>com.fasterxml.jackson.dataformat</groupId>
    <artifactId>jackson-dataformat-csv</artifactId>
    <version>2.14.1</version>
</dependency>
`

var basicDeps = fmt.Sprintf(basicJavaDepsTpl, "")
var updatedDeps = fmt.Sprintf(basicJavaDepsTpl, additionalLib)

const basicHandlerTpl = `
package io.project.kyma.serverless.handler;
        
        import jakarta.ws.rs.core.Context;
        import jakarta.ws.rs.core.Response;
        import io.project.kyma.serverless.sdk.CloudEvent;
        import io.project.kyma.serverless.sdk.Function;
        
        
        public class Handler implements Function {

            @Override
            public Response main(CloudEvent event, Context context) {
			return Response.ok("%s").build();
            }
        }
`

const csvHandler = `
package io.project.kyma.serverless.handler;
        import jakarta.ws.rs.core.Context;
        import jakarta.ws.rs.core.Response;
        import io.project.kyma.serverless.sdk.CloudEvent;
        import io.project.kyma.serverless.sdk.Function;
		import com.fasterxml.jackson.core.JsonProcessingException;
		import com.fasterxml.jackson.dataformat.csv.CsvMapper;
        
        
        public class Handler implements Function {

            @Override
            public Response main(CloudEvent event, Context context) {
				var msgData = new String[]{"Hello", "from", "new", "library"};
				var mapper = new CsvMapper();
				var msg = "empty";
				try {
					msg = mapper.writeValueAsString(msgData);
				} catch (JsonProcessingException e ){
					throw new RuntimeException(e);
				}
				return Response.ok(msg).build();
			}
        }


`

const ExpectedUpdatedResponse = `Hello,from,new,library`

func BasicJavaFunction(returnMsg string, rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	var spec = serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       fmt.Sprintf(basicHandlerTpl, returnMsg),
				Dependencies: basicDeps,
			},
		},
	}

	return spec
}

func UpdatedDepJavaCsvResponse(rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	var spec = serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       csvHandler,
				Dependencies: updatedDeps,
			},
		},
	}

	return spec
}
