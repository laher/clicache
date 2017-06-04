# clicache

Cache the STDOUT of a given command, for no more than a given duration (default is 5 minutes).

Call this command twice, and the second time will be super quick:

    clicache kubectl --context=dev get pod
    clicache kubectl --context=dev get pod

The initial call takes about 30 seconds, the second invocation takes milliseconds.

You can also delete entries with `-del`:

    clicache -del kubectl --context=dev get pod

For more options, use `clicache -h`
