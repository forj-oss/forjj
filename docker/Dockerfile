FROM fedora:23

MAINTAINER christophe.larsonneur@hpe.com

# This Dockerfile build the 'forjj-orchestrate' image to build/maintain your DevOps infrastructure

RUN dnf install wget git sudo tar jq -y && \
    dnf clean all

#ARG DOCKER_VERSION=latest

#RUN [ "$DOCKER_VERSION" != "" ] && \
#    wget -q -O /tmp/docker.tgz https://get.docker.io/builds/Linux/x86_64/docker-$DOCKER_VERSION.tgz && \
#    tar -xvzf /tmp/docker.tgz docker/docker && mv docker/docker /usr/bin/docker && \
#    rm -f /tmp/docker.tgz && rmdir docker

RUN useradd devops && \
    sed -i 's/^\(.*requiretty.*\)/\# \1/g' /etc/sudoers && \
    echo 'devops ALL=NOPASSWD:SETENV: /usr/bin/chown -R devops\:devops /devops' >> /etc/sudoers

VOLUME /devops 
RUN cat /etc/sudoers && mkdir /forjj-contribs #&& chown devops:devops /forjj-contribs
VOLUME /forjj-contribs

#COPY forjj/docker /usr/bin/docker
COPY forjj.sh /tmp/forjj.sh
COPY forjj/*.sh /usr/local/bin/

USER devops
ENTRYPOINT ["/tmp/forjj.sh"]
