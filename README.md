# Upgear's Go Kit

**Currently under heavy development: Make sure to vendor.**

This is a collection of Go libraries that we use on most projects at upgear.

We are building this kit in a pragmatic manner: As we start to notice repetition across our codebase, we consider moving the commonalities into this kit.

This library is opinionated in order to provide standardization, e.g: Logging is done via whitespace-separated key/value pairs (no JSON option).

**Design Goals:**

- Standardization over configuration
- Simplicity over all else (perfomance, versatility)

