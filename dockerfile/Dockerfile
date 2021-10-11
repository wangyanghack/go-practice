FROM centos
RUN yum install dlv -y \
  && yum install binutils -y \
  && yum install vim -y \
  && yum install gdb -y
RUN yum install -y wget
RUN wget https://golang.org/dl/go1.17.linux-amd64.tar.gz \
  && rm -rf /usr/local/go && tar -C /usr/local -xzf go1.17.linux-amd64.tar.gz
