package reload

/*
* Package reload provides support for detecting changes and reloading data files while the controller is running.
* In addition this package contains types that allow to "plug-into" controller logic in order to make such changes at runtime possible.
* It is necessary for things like SSL certificates that must be updated without downtime.
* For example: Without this feature, the controller Pod must be re-deployed after updating secrets with new SSL certificate data.
 */
