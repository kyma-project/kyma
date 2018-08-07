FROM eu.gcr.io/kyma-project/content-to-json-generator:0.0.1

ARG CONTENT_SRC_DIR
COPY ${CONTENT_SRC_DIR} /app/documentation

ARG APP_SRC_PATH
ARG APP_API_SRC_PATH
ARG APP_ASYNC_API_SRC_PATH
ARG APP_RESOURCES_BASE_URI
ARG APP_DOCS_CONFIG_PATH
ENV APP_OUT_PATH /out

RUN npm start

FROM eu.gcr.io/kyma-project/resources-uploader:0.0.1
COPY --from=0 /app/documentation/out /app/data
CMD "/app/main"