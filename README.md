## Docker-init

Docker-init is a lightweight init process mainly used in containers.

It does 4 things:
  - When starting, executes all the regular files in the given directory by passing the "start" argument
  - Listen for incoming interupt signals
  - When stopping, executes all the regular files in the given directory by passing the "stop" argument
  - Listen for incoming child exit signals to reap extra attached process

### Motivation

Yes I am using sysV init scripts in my containers. Those scripts usually dettach their children, which then gets attached to pid `1`.

When the same script is called again, the attached processes would stay in zombie state, this is where I needed something that could start scripts, stop scripts and cleanup the processes' states.

### Design choices

Create a directory in /etc/entrypoint with chmod `0700`.

Copy your shell scripts inside that will do the following:
  - if the target executable doesn't dettach, dettach
  - if the target executable doesn't drop privileges, drop them
  - call /bin/env to control what environment variable you pass to the target executable

### Usage

```
Usage of ./docker-init:
  -dir string
        Execute all the files from the directory
```

### Caveats

  - It doesn't filter scripts that are executed, so you should be careful what directory is specified
  - It starts processes using `root` so please, drop privileges in your executables
  - It doesn't dettach the executing process and *will* be blocked if so
  - It doesn't filter out env variables, so be careful to strip environment in your executables
