# YouTime

* Object to sample one NTP server and provide a TrueTime source
  * Collect coded samples, drop impure samples
  * Wait for N samples to be collected
  * Calculate skewPPB and skewRadiusPPB
  * Calculate ideal and radius (applying skew)
  * Serve intervals
* Compute joint interval
  * Get intervals from all N servers
  * Compute N intersection intervals by dropping one server for each
  * Compute union of intersection intervals as compound interval
    * `getRange(n relMoment) (earliest, latest time.Time)`
* Open questions:
  * Initial radii?
    * Coded samples ensure there is always a pair, is this enough?
  * What if joint interval doesn't exist? I.e. each server has a narrow time
  range but disjoint from other servers
    * Then you can't serve an interval. There does not exist a timestamp range
    that estimates the current time across the servers provided.
    * Drop servers that are persistently disjoint?
