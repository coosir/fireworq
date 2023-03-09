Make It Production-Ready
========================

- [Using a Release Build (Manual setup)][section-manual-setup]
- [Preparing a Backup Instance][section-backup]
- [Graceful Shutdown/Restart][section-graceful-restart]
- [Logging][section-logging]
- [Monitoring][section-monitoring]

## <a name="manual-setup">Using a Release Build (Manual Setup)</a>

A release build is available on [the releases page][releases].

For example, the following commands download and extract the Middleman
binary for Linux AMD64 (x86-64) platform.

```
$ OS=linux
$ ARCH=amd64
$ curl -L  $(curl -sL  https://api.github.com/repos/coosir/middleman/releases/latest | jq -r '.assets[].browser_download_url' | grep "_${OS}_${ARCH}.zip") > middleman_${OS}_${ARCH}.zip
$ unzip middleman_${OS}_${ARCH}.zip middleman
```

Before running Middleman, make sure that you have

<a name="manual-setup-mysql"></a>

1. [MySQL][] running somewhere in your host or network on
   <code><var>mysql_host</var>:<var>mysql_port</var></code>,
1. some database <code><var>database</var></code> on the MySQL server,
   and,
1. some DB user <code><var>user</var></code> with some password
   <code><var>password</var></code>, who is granted `CREATE`,
   `INSERT`, `DELETE`, `UPDATE` and `SELECT` rights on
   <code><var>database</var></code>.

Then the following commands run Middleman with the prepared MySQL database.

<pre><code>
$ export MIDDLEMAN_MYSQL_DSN=<var>user</var>:<var>password</var>@tcp(<var>mysql_host</var>:<var>mysql_port</var>)/<var>database</var>
$ export MIDDLEMAN_QUEUE_DEFAULT=default
$ export MIDDLEMAN_BIND=0.0.0.0:8080
$ ./middleman
</code></pre>

You can specify different MySQL database names, hosts, users or
passwords for <code>MIDDLEMAN_REPOSITORY_MYSQL_DSN</code> and
<code>MIDDLEMAN_QUEUE_MYSQL_DSN</code> if you prefer.

## <a name="backup">Preparing a Backup Instance</a>

Middleman provides a mechanism to run a fail-safe backup instance for
redundancy.  You have only to run a secondary instance against the
same DB configuration to achieve this.  You can even run more than two
instances.  Those instances are typically on a different host from
each other.

When multiple instances get ready, whichever instance accepts
[pushing a job][api-post-job], which will be handled by an active
instance for the queue.  When the active instance dies, another
instance will be active automatically.  If the underlying DB server
dies, all the instances get inactive until the DB server recovers.

Note that multiple instances theoretically form a cluster; each queue
may be handled by a different instance.  This situation is unlikely to
happen for now because there is no way to deactivate a single queue
handling, but there is no guarantee that an active instance handles
all the queues and the others are totally inactive.

## <a name="graceful-restart">Graceful Shutdown/Restart</a>

### Shutdown

A Middleman daemon can be terminated gracefully by `SIGINT`, `SIGTERM`
or `SIGHUP`.  It will wait for accepted API requests to be processed
and grabbed jobs to be completed until timeout specified by
[`MIDDLEMAN_SHUTDOWN_TIMEOUT`][env-shutdown-timeout] occurs.

### Restart

To restart a Middleman daemon gracefully, wrap the daemon with a tool
like [`start_server`][start_server].

```
$ export ...
$ start_server --port=8080 -- ./middleman
```

Sending `SIGTERM` or `SIGHUP` to the `start_server` process will
gracefully shutdown or restart the daemon respectively.

## <a name="logging">Logging</a>

Middleman has three types of logs: an error log, an access log and a
queue log.  The error log goes to the standard error in a colorized
pretty format and the access log and the queue log goes to the
standard output in JSON format by default.  You can change the
destinations of log outputs to files by specifying
[`MIDDLEMAN_ERROR_LOG`][env-error-log],
[`MIDDLEMAN_ACCESS_LOG`][env-access-log] and
[`MIDDLEMAN_QUEUE_LOG`][env-queue-log] respectively.  The error log
will also be in JSON format if it is written to a file.  You can
control what to output to the error log and the queue log by
specifying [`MIDDLEMAN_ERROR_LOG_LEVEL`][env-error-log-level] and
[`MIDDLEMAN_QUEUE_LOG_LEVEL`][env-queue-log-level] respectively.

To rotate the file logs, use a tool like [`logrotate`][logrotate].  Be
aware that some log lines may be lost by `logrotate` with
`copytruncate`.  Instead, ask Middleman to reopen the log files by
sending `USR1` signal after a rotation.  It should be something like
this:

```
/var/log/middleman/*.log {
  rotate 7
  size 10k
  missingok
  notifempty
  sharedscripts
  postrotate
    [ -f /var/run/middleman.pid ] && kill -USR1 `cat /var/run/middleman.pid`
  endscript
}
```

Note that you need [`MIDDLEMAN_PID=/var/run/middleman.pid`][env-pid] to
get it work.

## <a name="monitoring">Monitoring</a>

Middleman provides some statistics and they can be monitored by your
favorite monitoring tool such as [Mackerel][], [Zabbix][], [Sensu][]
or [Munin][].

### Go Stats

Statistics of Go runtime metrics are provided by `/stats`.  You can
easily monitor these metrics by using plugins for your monitoring
tool.

|Tool    |Plugin                                                  |
|--------|--------------------------------------------------------|
|Zabbix  |[fukata/golang-stats-api-handler-zabbix-userparameter][]|
|Sensu   |[sensu-plugins-golang][]                                |
|Munin   |[fukata/golang-stats-api-handler-munin-plugin][]        |

### Queue Stats

Statistics of job queue metrics are provided by `/queues/stats` or
<code>/queue/<var>{queue_name}</var>/stats</code>.  

### Alerts

You can get alerts when a job permanently failed by using your monitoring tool.


[section-manual-setup]: #manual-setup
[section-backup]: #backup
[section-graceful-restart]: #graceful-restart
[section-logging]: #logging
[section-monitoring]: #monitoring
[api-post-job]: ./api.md#api-post-job

[env-access-log]: ./config.md#env-access-log
[env-error-log]: ./config.md#env-error-log
[env-error-log-level]: ./config.md#env-error-log-level
[env-queue-log]: ./config.md#env-queue-log
[env-queue-log-level]: ./config.md#env-queue-log-level
[env-pid]: ./config.md#env-pid
[env-shutdown-timeout]: ./config.md#env-shutdown-timeout

[releases]: https://github.com/coosir/middleman/releases

[Docker]: https://www.docker.com/
[MySQL]: https://www.mysql.com/
[start_server]: https://metacpan.org/pod/distribution/Server-Starter/script/start_server
[logrotate]: https://github.com/logrotate/logrotate
[Zabbix]: https://www.zabbix.com/
[Sensu]: https://sensuapp.org/
[Munin]: http://munin-monitoring.org/

[fukata/golang-stats-api-handler-zabbix-userparameter]: https://github.com/fukata/golang-stats-api-handler-zabbix-userparameter
[sensu-plugins-golang]: https://github.com/sensu-plugins/sensu-plugins-golang/blob/master/bin/metrics-golang-stats-api.rb
[fukata/golang-stats-api-handler-munin-plugin]: https://github.com/fukata/golang-stats-api-handler-munin-plugin
