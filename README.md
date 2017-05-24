# Upgear's Go Kit

This is a collection of Go libraries that we use on most projects at upgear.

We are building this kit in a pragmatic manner: As we start to notice repetition across our codebase, we consider moving the commonalities into this kit.

**This library is opinionated in order to provide standardization, e.g:**

Logging: Whitespace-separated key/value pairs (i.e: `a=b c=d`)
Configuration: Environment variables

Design Goals:
- Standardization over configuration
- Simplicity over all else (perfomance, versatility)

