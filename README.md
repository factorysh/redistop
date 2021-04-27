Redis Top
=========

Redistop uses [MONITOR](https://redis.io/commands/monitor) to watch redis commands.

> Because MONITOR streams back all commands, its use comes at a cost.

Test
----

    make
    ./redistop localhost:6379
