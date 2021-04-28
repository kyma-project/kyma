---
title: Debugging the Kyma installation
type: Troubleshooting
---

Use the alpha commands for error handling, for example:

- To get a detailed view of the installation process, use the `--verbose` flag.
- To tweak the values on a component level, use `alpha deploy --components`: Pass a components list that includes only the components you want to test and try out the settings that work for your installation.
- To understand which component failed during deployment, *deactivate* the default atomic deployment: `--atomic=false`. 
   With atomic deployment active, any component that hasn't been installed successfully is rolled back, which may make it hard to find out what went wrong. By disabling the flag, the failed components are not rolled back.

<!-- ANY OTHER DEBUGGING USE CASES? -->
