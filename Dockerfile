FROM centos:7

ENV container docker
RUN (sed -e 's|^mirrorlist=|#mirrorlist=|g' \
    -e 's|^#baseurl=http://mirror.centos.org/centos|baseurl=https://mirrors.tuna.tsinghua.edu.cn/centos|g' \
    -i.bak \
    /etc/yum.repos.d/CentOS-*.repo \
    )
RUN (yum clean all && \
    yum makecache && \
    yum update -y && \
    yum install -y binutils vim gdb wget openssh-server && \
    echo "root:12345678" | chpasswd && \
    ssh-keygen -t dsa -f /etc/ssh/ssh_host_dsa_key && \
    ssh-keygen -t rsa -f /etc/ssh/ssh_host_rsa_key && \
    ssh-keygen -t ed25519 -f /etc/ssh/ssh_host_ed25519_key && \
    ssh-keygen -t ecdsa -f /etc/ssh/ssh_host_ecdsa_key && \
    mkdir -p /var/run/sshd && \
    mkdir -p /home/admin/.ssh && \
    sed -ri 's/session reqired pam_loginuid.so/#session requied pam_loginuid.so/g' /etc/pam.d/sshd && \
    wget https://golang.google.cn/dl/go1.20.1.linux-amd64.tar.gz && \
    tar -zxvf go1.20.1.linux-amd64.tar.gz -C /usr/lib/ && \
    rm -rf /etc/ssh/sshd_config && \
    rm -rf /root/.bash_profile \
    )

EXPOSE 22 80 1234

RUN (yum install -y gcc)

RUN (mkdir -p /home/mit-6.824-lab1)
ADD  . /home/mit-6.824-lab1

ENV PATH $PATH:/usr/lib/go/bin

ADD .bash_profile /root/
ADD ssh* /etc/ssh/
RUN (source /root/.bash_profile && \
    go env -w GO111MODULE=on && \
    go env -w GOPROXY=https://goproxy.cn,direct && \
    go env -w GOSUMDB=sum.golang.google.cn && \
    cd /home/mit-6.824-lab1/src/ && \
    go get gopkg.in/yaml.v3 && \
    go mod tidy \
    )

