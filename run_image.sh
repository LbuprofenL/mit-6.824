docker build -t centos-godev:v1 .
docker run -d --privileged --name 6.824 -p 18080:22  centos-godev:v1 /usr/sbin/sshd -D &