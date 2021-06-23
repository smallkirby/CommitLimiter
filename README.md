# Commit Limiter: Prohibit Commits to Github
Yes, I have what I have to do.  
Commits to Github is becoming a way of **Escapism** for me.  
Ok, prohibit Github, and commit to my life...

## Installation
```install.sh
git clone https://github.com/smallkirby/CommitLimiter.git
cd ./CommitLimiter
make install
smgithub --init --username <YOUR USERNAME> --limit <NUM>
```

## Usage

- Follow instruction in [Installation](#installation).  
  - This would install the binary in `/usr/bin/smgithub` and create configuration file in `/etc/smgithub/setting.conf`.  
  - Also, it registers a cron task at `/etc/cron.d/smgithub`.  

- It automatically checks your Github activity every hours, then prohibit more commits after it exceeds threshold.

## Progress
| Status | Functionality |
| ------------- | ------------- |
| ☀️ | fetch commits |
| ☀️ | impl threshold |
| ☀️ | prohibit commits |

### legend
- ☀️: completed
- 🌤: almost done, still needs more impls 
- ☁️: work in progress
- ⛈: totally untouched

## Warnings

- This program is intended to be run as root as crontask, and it overwrites `/etc/hosts` directly without any security and sync issues. Be aware.
