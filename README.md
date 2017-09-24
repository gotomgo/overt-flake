# overt-flake
Flake ID Generation server developed for use with Overtone - by Overtone Studios

Overt-flake is a Flake ID generator and server (written in GO) along the lines of Twitter Snowflake, Boundary Flake and others. It deviates from other implementations in small but important ways:

1. Identifiers are 128-bits
2. External configuration information such as worker id and data-center id are not needed. Machine identifiers, both stable and unstable are used instead
3. The Overtone Epoch (Jan 1, 2017) is used, rather than Twitter Epoch or Unix Epoch. The primary reason is that 1/1/2017 is a Sunday (not important for ID generation) and nostalgically, it is also the day technical development of Overtone began.
