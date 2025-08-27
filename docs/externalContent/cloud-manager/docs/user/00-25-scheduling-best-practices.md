# Scheduling Best Practices
<!--Scheduling Best Practices are not part of Help Portal docs-->

> [!WARNING]
> NFS backup scheduling is a beta feature available only per request for SAP-internal teams.

To create backups at specified intervals, use the NFS backup scheduling. To use the feature effectively, follow the scheduling best practices:

* Configure the `MaxRetentionDays`, `MaxReadyBackups`, and `MaxFailedBackups` attributes on the schedule to auto-delete the oldest backups when any of these thresholds are exceeded.
* Create multiple backup schedules with different frequencies to reduce the number of backups and increase the coverage of the backup period. For example, to create a backup plan for 10 years, you can configure some or all of the following schedules:
  * `hourly schedule with max retention period of 1 day`
  * `daily schedule with max retention period of 7 days`
  * `weekly schedule with max retention period of 35 days`
  * `monthly schedule with max retention period of 365 days`
  * `yearly schedule with max retention period of 3650 days`
* Contact the SRE team to increase the cloud provider quota limits.
