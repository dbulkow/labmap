FROM registry.access.redhat.com/rhel7:7.2-84
MAINTAINER David Bulkow <david.bulkow@stratus.com>

ADD labmap /usr/bin

CMD ["labmap"]
