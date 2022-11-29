FROM ${BASE_IMAGE} as builder

ARG BUILD_DIR=/build
ARG SOURCE_DIR=/src
ARG DEPS_DIR=/src
WORKDIR $BUILD_DIR

COPY $DEPS_DIR/pom.xml $BUILD_DIR/handler-pom.xml
#TODO: handler.java is in root not in src
COPY $SOURCE_DIR/Handler.java $BUILD_DIR/src/main/java/io/project/kyma/serverless/handler/Handler.java
RUN cat /build/src/main/java/io/project/kyma/serverless/handler/Handler.java
RUN mvn dependency:go-offline -f handler-pom.xml

#TODO: Reconsider: I delete this Handler class to compile only the new Handler class
RUN mvn clean && mvn package -f pom.xml

FROM 192.168.122.1:5000/openjdk:11-jre

COPY --from=builder /build/target/kyma-java-runtime-0.0.1.jar /app.jar

ENTRYPOINT java -Djava.security.egd=file:/dev/./urandom -jar /app.jar
USER 1000