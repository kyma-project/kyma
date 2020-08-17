# Git Server

This is a tool that exposes an HTTP server with git repository, that can be used in e2e tests, to avoid connections with external git providers.

Additional git repositories can be added simply by creating new directory in `/repos/`.
The name of the created directory will be used as the repository name.  
