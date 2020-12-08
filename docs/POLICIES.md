## Writing runtime policies

The policy engine adopts and extends the Falco rules definition syntax. Before reading the rest of this section, please go through the [Falco Rules](https://falco.org/docs/rules/) documentation to get familiar with _rule_, _macro_, and _list_ syntax, all of which are supported in our policy engine. Policies are written in one or more `yaml` files, and stored in a directory specified in the pipeline configuration file under the `policies` attribute of the policy engine plugin.  

*Rules* contain the following fields:

- _rule_: the name of the rule
- _description_: a textual description of the rule
- _condition_: a set of logical operations that can reference lists and macros, which when evaluating to _true_, can trigger an alert (when processor is in alert mode), or filter a sysflow record (when processor is in filter mode)
- _action_: a list of actions to take place when the rule evaluates to _true_. Actions can be any of the following (note: new actions will be added in the future):
  - alert: processor outputs an alert
  - tag: enriches or tags the sysflow record with the labels in the `tags` field. This can be useful for semantically labeling of records with TTPs for example.
- _priority_: label representing the severity of the alert can be: (1) low, medium, or high, or (2) emergency, alert, critical, error, warning, notice, informational, debug.
- _tags_ (optional): set of labels appended to alert (default: empty).
- _prefilter_ (optional): list of record types (`sf.type`) to whitelist before applying rule condition (default: empty).
- _enabled_ (optional): indicates whether the rule is enabled (default: true).

*Macros* are named conditions and contain the following fields:

- _macro_: the name of the macro
- _condition_: a set of logical operations that can reference lists and macros, which evaluate to _true_ or _false_

*Lists* are named collections and contain the following fields:

- _list_: the name of the list
- _items_: a collection of values or lists

*Filters* blacklist records matching a condition:

- _filter_: the name of the filter
- _condition_: a set of logical operations that can reference lists and macros, which evaluate to _true_ or _false_

For example, the rule below specifies that matching records are process events (`sf.type = PE`), denoting `EXEC` operations (`sf.opflags = EXEC`) for which the process matches macro `package_installers`. Additionally, the gloabl filter `containers` preempitively removes from the processing stream any records for processes not running in a container environment.

```yaml
# lists
- list: rpm_binaries
  items: [dnf, rpm, rpmkey, yum, '"75-system-updat"', rhsmcertd-worke, subscription-ma,
          repoquery, rpmkeys, rpmq, yum-cron, yum-config-mana, yum-debug-dump,
          abrt-action-sav, rpmdb_stat, microdnf, rhn_check, yumdb]

- list: deb_binaries
  items: [dpkg, dpkg-preconfigu, dpkg-reconfigur, dpkg-divert, apt, apt-get, aptitude,
    frontend, preinst, add-apt-reposit, apt-auto-remova, apt-key,
    apt-listchanges, unattended-upgr, apt-add-reposit]

- list: package_mgmt_binaries
  items: [rpm_binaries, deb_binaries, update-alternat, gem, pip, pip3, sane-utils.post, alternatives, chef-client]

# macros
- macro: package_installers
  condition: sf.proc.name pmatch (package_mgmt_binaries)

# global filters (blacklisting)
- filter: containers
  condition: sf.container.type = host

# rule definitions
- rule: Package installer detected
  desc: Use of package installer detected
  condition:  sf.opflags = EXEC and package_installers
  action: [alert]
  priority: medium
  tags: [actionable-offense, suspicious-process]
  prefilter: [PE] # record types for which this rule should be applied (whitelisting)
  enabled: true
```

The following table shows a detailed list of attribute names supported by the policy engine, as well as their
type, and comparative Falco attribute name. Our policy engine supports both SysFlow and Falco attribute naming convention to enable reuse of policies across the two frameworks.

| Attributes     | Description       | Values | Falco Attribute |
|:----------------|:-----------------|:------|----------|
| sf.type           | Record type       | PE,PF,NF,FF,FE | N/A |
| sf.opflags        | Operation flags   | [Operation Flags List](https://sysflow.readthedocs.io/en/latest/spec.html#operation-flags): remove `OP_` prefix | evt.type (remapped as falco event types) |
| sf.ret            | Return code       | int   |  evt.res |
| sf.ts             | start timestamp(ns)| int64 | evt.time |
| sf.endts          | end timestamp(ns) | int64  |  N/A |
| sf.proc.pid       | Process PID       | int64  | proc.pid |
| sf.proc.tid       | Thread PID        | int64  | thread.tid |
| sf.proc.uid       | Process user ID   | int    | user.uid |
| sf.proc.user      | Process user name | string | user.name |
| sf.proc.gid       | Process group ID  | int    | group.gid |
| sf.proc.group     | Process group name | string | group.name |
| sf.proc.apid      | Proc ancestors PIDs (qo) | int64 | proc.apid |
| sf.proc.aname     | Proc anctrs names (qo) (exclude path) | string | proc.aname |
| sf.proc.exe       | Process command/filename (with path) | string | proc.exe |
| sf.proc.args      | Process command arguments | string | proc.args |
| sf.proc.name      | Process name (qo) (exclude path) | string | proc.name |
| sf.proc.cmdline   | Process command line (qo) | string | proc.cmdline |
| sf.proc.tty       | Process TTY status | boolean | proc.tty |
| sf.proc.entry     | Process container entrypoint | bool |  proc.vpid == 1 |
| sf.proc.createts  | Process creation timestamp (ns) | int64 | N/A |
| sf.pproc.pid      | Parent process ID | int64 | proc.ppid |
| sf.pproc.gid      | Parent process group ID | int64 | N/A |
| sf.pproc.uid      | Parent process user ID  | int64 | N/A |
| sf.pproc.group    | Parent process group name | string | N/A |
| sf.pproc.tty      | Parent process TTY status | bool | N/A |
| sf.pproc.entry    | Parent process container entry | bool | N/A |
| sf.pproc.user     | Parent process user name | string | N/A |
| sf.pproc.exe      | Parent process command/filename | string | N/A |
| sf.pproc.args     | Parent process command arguments | string | N/A |
| sf.pproc.name     | Parent process name (qo) (no path) | string | proc.pname |
| sf.pproc.cmdline  | Parent process command line (qo) | string | proc.pcmdline |
| sf.pproc.createts | Parent process creation timestamp | int64 | N/A |
| sf.file.fd        | File descriptor number | int |  fd.num |
| sf.file.path      | File path | string | fd.name |
| sf.file.newpath   | New file path (used in some FileEvents) | string | N/A |
| sf.file.name      | File name (qo) | string | fd.filename |
| sf.file.directory | File directory (qo) | string | fd.directory |
| sf.file.type      | File type | char 'f': file, 4: IPv4, 6: IPv6, 'u': unix socket, 'p': pipe, 'e': eventfd, 's': signalfd, 'l': eventpoll, 'i': inotify, 'o': unknown. | fd.typechar |  
| sf.file.is_open_write | File open with write flag (qo) | bool | evt.is_open_write |
| sf.file.is_open_read | File open with read flag (qo) | bool | evt.is_open_read |
| sf.file.openflags | File open flags | int | evt.args |
| sf.net.proto      | Network protocol | int | fd.l4proto |
| sf.net.sport      | Source port  | int | fd.sport |
| sf.net.dport      | Destination port | int | fd.dport |
| sf.net.port       | Src or Dst port (qo) | int | fd.port |
| sf.net.sip        | Source IP | int  | fd.sip |
| sf.net.dip        | Destination IP | int | fd.dip |
| sf.net.ip         | Src or dst IP (qo) | int | fd.ip |
| sf.res            | File or network resource | string | fd.name |
| sf.flow.rbytes    | Flow bytes read/received | int64 |  evt.res |
| sf.flow.rops      | Flow operations read/received | int64 | N/A |
| sf.flow.wbytes    | Flow bytes written/sent | int64 | evt.res |
| sf.flow.wops      | Flow bytes written/sent | int64 | N/A |
| sf.container.id   | Container ID | string | container.id |
| sf.container.name | Container name | string | container.name |
| sf.container.image.id | Container image ID | string | container.image.id |
| sf.container.image | Container image name  | string | container.image |
| sf.container.type | Container type | CT_DOCKER, CT_LXC, CT_LIBVIRT_LXC, CT_MESOS, CT_RKT, CT_CUSTOM, CT_CRI, CT_CONTAINERD, CT_CRIO, CT_BPM | container.type |
| sf.container.privileged | Container privilege status | bool | container.privileged |
| sf.node.id        | Node identifier | string |  N/A |
| sf.node.ip        | Node IP address | string | N/A |
| sf.schema.version | SysFlow schema version | string | N/A |
| sf.version        | SysFlow JSON schema version  | int | N/A |

The policy language supports the following operations:

| Operation | Description | Example |
|:----------|:------------|:--------|
| A and B | Returns true if both statements are true | sf.pproc.name=bash and sf.pproc.cmdline contains echo |
| A or B | Returns true if one of the statements are true | sf.file.path = "/etc/passwd" or sf.file.path = "/etc/shadow" |
| not A | Returns true if the statement isn't true | not sf.pproc.exe = /usr/local/sbin/runc | 
| A = B| Returns true if A exactly matches B.  Note, if B is a list, A only has to exact match one element of the list.  If B is a list, it must be explicit.  It cannot be a variable.  If B is a variable use `in` instead. | sf.file.path = ["/etc/passwd", "/etc/shadow"] |
| A != B| Returns true if A is not equal to B.  Note, if B is a list, A only has to be not equal to one element of the list. If B is a list, it must be explicit.  It cannot be a variable. | sf.file.path != "/etc/passwd"|
| A < B |  Returns true if A is less than B.  Note, if B is a list, A only has to be less than one element in the list. If B is a list, it must be explicit.  It cannot be a variable. | sf.flow.wops < 1000 |
| A <= B |  Returns true if A is less than or equal to B.  Note, if B is a list, A only has to be less than or equal to one element in the list. If B is a list, it must be explicit.  It cannot be a variable. | sf.flow.wops <= 1000 | 
| A > B |  Returns true if A is greater than B.  Note, if B is a list, A only has to be greater than one element in the list. If B is a list, it must be explicit.  It cannot be a variable. | sf.flow.wops > 1000 |
| A >= B |  Returns true if A is greater than or equal to B.  Note, if B is a list, A only has to be greater than or equal to one element in the list. If B is a list, it must be explicit.  It cannot be a variable. | sf.flow.wops >= 1000 |
| A in B |  Returns true if value A is an exact match to one of the elements in list B. Note: B must be a list.  Note: () can be used on B to merge multiple list objects into one list. | sf.proc.exe in (bin_binaries, usr_bin_binaries) |
| A startswith B | Returns true if string A starts with string B |  sf.file.path startswith '/home' |
| A endswith B | Returns true if string A ends with string B |  sf.file.path endswith '.json' |
| A contains B |  Returns true if string A contains string B |  sf.pproc.name=java and sf.pproc.cmdline contains org.apache.hadoop |
| A icontains B |  Returns true if string A contains string B ignoring capitalization |  sf.pproc.name=java and sf.pproc.cmdline icontains org.apache.hadooP |
| A pmatch B |  Returns true if string A partial matches one of the elements in B. Note: B must be a list.  Note: () can be used on B to merge multiple list objects into one list. |  sf.proc.name pmatch (modify_passwd_binaries, verify_passwd_binaries, user_util_binaries) |
| exists A | Checks if A is not a zero value (i.e. 0 for int, "" for string)|  exists sf.file.path |

See the resources policies directory in [github](https://github.com/sysflow-telemetry/sf-processor/tree/master/resources/policies) for examples. Feel free to contribute new and interesting rules through a github pull request.
