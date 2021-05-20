Redis Top
=========

Redistop uses [MONITOR](https://redis.io/commands/monitor) to watch Redis
commands and shows per command and per host statistics.

> Because MONITOR streams back all commands, its use comes at a cost.

Redistop uses INFO command too.

Example
-------

![Redis Top screenshot](redistop.png)

Build
-----

If you have recent golang dev enironment set, you can build it with the Makefile

    make

If you need a Linux compilation, or juste using Docker:

    make docker-build

License
-------

GPL v3
