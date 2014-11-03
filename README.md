Code Runner
===========

An application to run code

Installation:
=============
- `brew install boot2docker`
- added runner to path

Usage:
=====
    NAME:
       Runner - Run some code...

    USAGE:
       Runner [global options] command [command options] [arguments...]

    VERSION:
       0.0.1a

    COMMANDS:
       help, h  Shows a list of commands or help for one command
       
    GLOBAL OPTIONS:
       --lang, -l 'golang'  language for runner
       --help, -h   show help
       --version, -v  print the version

Languages Natively Supported:
====================

- Go: -l golang
- Node.js: -l node

Configuration
==========================

## Languages
Adding languages to code-runner is easy just create a `~/.coderunner/config.yml`
and add the following to it:

    languages:
      ruby:
        command: "ruby %s"
        container: ruby

you can also set a default language by adding
    
    default: yourlang


