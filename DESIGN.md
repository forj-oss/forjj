# Introduction

This file describes how Forjj core works.

In short, the core read the Forjfile, load plugins and flow, define
extra cli parameters (plugin/flow), apply flow and run actions.

Because we use kingpin cli and plugins/flows can define dynamically
additional parameters, we have at least 2 big steps.

One during cli setup, driven by ParseContext.
one after cli setup driven by actions.

# Details

## ParseContext

1. Forfile loaded
2. Paths setup from cli or ENV or Forjfile or default.
3. Load plugins
4. Default upstream to all repos except those already set.
5. flow load and apply

## Cli setup

This sequence is complete by forjj-cli

## after cli

Mainly the core is made in a function called `ScanAndSetObjects`

1. dispatch to creds for secure data
2. Set defaults on all plugins objects instances.

