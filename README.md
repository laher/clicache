# clicache

Cache the STDOUT of a given command, for no more than a given duration (default is 5 minutes).

_NOTE: I created this with `kubectl get pod` in mind. It's not specific to `kubectl` (part of the kubernetes project) but it would need some more work to become more generally useful._

NOTE: This is totally experimental. Only use it for something with smallish output, and be careful with write commands.

## Install

    go get -u github.com/laher/clicache

## Overview

Call this command twice, and the second time will be super quick:

    clicache kubectl --context=dev get pod
    clicache kubectl --context=dev get pod

The initial call takes about 30 seconds, the second invocation takes milliseconds.

You can also delete entries with `-del`:

    clicache -del kubectl --context=dev get pod

For more options, use `clicache -h`


## NOTE on kubectl

I include the following functions in my zsh config (hopefully it works in bash too)

```

function ki {
  clicache kubectl --context=$1 get pod | grep $2 | cut -d ' ' -f1
}

function k {
  kubectl --context=$1 ${@:2}
}

function kclc {
  clicache -del kubectl --context=$1 get pod
}

```

Now, given a context `dev` and a pod named `mypod` I can quickly run commands like this (without clicache there would be a few seconds delay fetching the container name):

    k dev log -f $(ki dev mypod)
    k dev exec -it $(ki dev mypod) sh

In the future I'd like to wrap `kclc dev` into some other commands, in order to explicitly clear cache
