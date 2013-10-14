Ambari pre-requests
===

Configures given cluster of machines to be ready to setup [Apache Ambari](http://incubator.apache.org/ambari/)

Cross-compile
=====
To build it for CentOS 6 use
```
GOOS=linux GOARCH=amd64 go build -o ambari-prereq-linux64 github.com/alexander-bzz/ambari-prereq/
```
