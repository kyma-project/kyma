# JobInit

The JobInit is used as the init container for injecting git repository to the [building job](https://kyma-project.io/docs/components/serverless/#details-function-processing-built).

### Environment variables

JobInit uses these environment variables:

| Variable                    | Description                                                                   | Default value |
| --------------------------- | ----------------------------------------------------------------------------- | ------------- |
| **APP_MOUNT_PATH**          | Destination path on which should be cloned the repository                     | `/workspace`
| **APP_REPOSITORY_URL**      | Address to the git repository on which should be cloned                       |
| **APP_REPOSITORY_COMMIT**   | Commit Hash on which should be checkout                                       |
| **APP_REPOSITORY_USERNAME** | Username of account which should be used to clone private repository          |
| **APP_REPOSITORY_PASSWORD** | Password or token of account which should be used to clone private repository |
