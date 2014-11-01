Code Runner
===========

An application to run code

Installation:
=============
- `brew install boot2docker`
- added runner to path

Usage:
=====
    runner [-h, -l -c|--command, --container] [FILE]
    Run code and edit it in the same window...
        
        -h                  display this help and exit
        -l                  Language to use
        -c, --command       Command to run in the contianer
        -i                  Docker Container to use

Languages Supported:
====================

- Go: -l golang
- Node.js: -l node
- Python: -l python
- Ruby: -l ruby