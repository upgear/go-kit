# Upgear's Go Kit

This is a collection of Go libraries that we use on most projects at upgear.

The goal is not to provide a library that can be used in any environment. It is
to provide the common utilities which can be used in our environment 95% of the
time. It is built in a pragmatic manner: As we notice that we write the same
code over and over, we consider moving it into this kit.

This library is opinionated in order to provide standardization, e.g:**

Logging: Whitespace-separated key/value pairs (i.e: `a=b c=d`)
Configuration: Environment variables

Design Goals:
- Standardization over configuration
- Simplicity over all else (perfomance, versatility)

