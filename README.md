# backive

The name comes from the combination of backup and archive - silly, I know.

# Purpose

I've a long-term backup strategy where I put some labeled hard-disk in a SATA docking station and run a backup routine. When done, this hard-disk goes back into some shelf in my attic or basement. When the time is come again to update the long-term backup the same procedure starts again.

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

# Current state

Working daemon and udev binary. Ready for first basic usage and testing in "production".
