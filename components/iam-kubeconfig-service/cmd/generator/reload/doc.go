package reload

/*
* Package reload provides support for detecting changes to data files while the controller is running.
* In addition this package contains types for plugging into controller logic in order to reload the data from the files once changes are detected.
* These features are necessary for things like SSL certificates that must be updated at runtime without downtime.
* Otherwise, the controller Pod must be re-deployed after updating secrets with new SSL certificate data.
* The file-watch facilites are separated from data-reloading components for better extensibility and reusability.
 */
