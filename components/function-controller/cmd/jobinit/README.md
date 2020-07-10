# JobInit

The JobInit is used as the init container for injecting a Git repository to the [Job that builds a Function](https://kyma-project.io/docs/components/serverless/#details-function-processing-built).

### Environment variables

JobInit uses these environment variables:

| Variable                    | Description                                                                   | Default value |
| --------------------------- | ----------------------------------------------------------------------------- | ------------- |
| **APP_MOUNT_PATH**          | Path under which JobInit should clone the repository                     | `/workspace`
| **APP_REPOSITORY_URL**      | Address of the Git repository to clone                       |
| **APP_REPOSITORY_COMMIT**   | Commit to check out when cloning the repository                                   |
| **APP_REPOSITORY_USERNAME** | Username of the account which should be used to clone the private repository          |
| **APP_REPOSITORY_PASSWORD** | Password or token for the account which should be used to clone the private repository |
