# Backive

The name comes from the combination of backup and archive - silly, I know.

# The initial idea
## Purpose

I have a long-term backup strategy where I put some labeled hard-disk in a SATA docking station and run a backup routine. When done, this hard-disk goes back into some shelf in my attic or basement. When the time is come again to update the long-term backup the same procedure starts again.

So now there are my backup routines, which are manually currently - **and that sucks.**

So what this tool or service should do is the following:
- I am able to configure based on UUIDs of the partitions which devices are allowed for backup
- I can specify as much backup items as I want, which should include:
  - Backup local and remote data (Linux machine and SSH required)
  - Use the best tool available for the task (rsync, duplicity, whatever)
  - Even be able to "backup" without a target device (choose another path on the system)
  - (optional) Be able to run pre-backup commands (for databases maybe), remote too
- The service has to be able to automatically detect the presence of a hard-disk, mounting it, running the backup routine and unmounting
- Good logging about the process


What I currently see as optional:
- Notification about the finished process

## Technical goals

- systemd service
- udev rules for notifications about new drives
- Golang implementation


# Reached goals

## Configuration

Here you see an example configuration, should be rather self explanatory.

```yaml
devices:
  dev_test:
    uuid: 714dc1f8-00d7-44b4-b03c-b54498c7cb86
    owner: myuser
  MyFATDevice:
    uuid: 72A9-D8C5
    owner: myuser
backups:
  dev_test_backup:
    user: myuser
    targetDevice: dev_test
    frequency: 2
    scriptPath: /home/myuser/backup.sh
    sourcePath: /home/myuser/data2
    targetPath: data2
    label: "Development test backup"
  scanner_usbstick_test:
    user: myuser
    targetDevice: MyFATDevice
    frequency: 0
    scriptPath: /absolute/path/to/backup.sh
    sourcePath: /home/myuser/data
    targetPath: targetdata
settings:
  systemMountPoint: /mnt/backups
  unixSocketLocation: /var/local/backive/backive.sock
  logLocation: /var/log/backive
  dbLocation: /var/lib/backive
```
## Backup scripts

If the `user` in a backup configuration is NOT defined, the scripts are executed as **root**!
This might change in the future, that execution as root is not allowed anymore.

The scripts awaited in the backup config string `scriptPath` are executed with the user defined in `user`.
The `sh` shell is used, so if you want to switch, do not forget `#!/usr/bin/env <yourshell>`.
The `scriptPath` should be absolute and should return a `0` on success.

What and how you do your backup with the script is completely up to you.

You can rely on following environment variables in your script:
```bash
BACKIVE_MOUNT     #The system mount point, where backive mounts it's known devices.
BACKIVE_TO        #The target directory as absolute path.
BACKIVE_FROM      #The source directory as absolute path.
```
If your backup script does not need the `BACKIVE_FROM` you currently still have to provide a path, just use a dummy.

## systemd service

**Has still to be done manually.**

- So if you want to test backive, copy the file `udev/99-backive.rules` to `/etc/udev/rules.d/` and change the path to the `backive_udev` binary.
- Do not forget to give all binaries a `chmod +x`
- Put the `systemd/backive.service` to `/etc/systemd/system/` and do a systemctl daemon-reload
  - The service needs root rights to run, sadly, to be able to use the specified users in the configuration for executing, I know that this is a security issue, but because of the fact it is currently only a service for myself I do not care. It might be changed later to a group permission management.
- Run the service and keep an eye on `/var/log/backive/` and it's files there

# Things to do

## Creating an install script

A script from the repo or in the "distribution" to setup with standard expectations (see systemd service).

## Increasing security

Move from root execution to a `backive` user and `backive` group. Users who want to backup stuff, need to be in the group. And the permissions for the data will be given to the user and the `backive` group.

Creation of this group and the handling should also be done with the install script.

## GUI for the icon bar

A GUI is planned, which notifies you of due backups you've defined, based on the last backup date.

An extension of the GUI would be an editor for the configuration or even a "wizard" dialog to add new backups or devices. This will only work if the group handling is implemented.

# Final words

## Thanks

To myself, finally having done this. lol...

Golang, it's awesome and nice to learn.
