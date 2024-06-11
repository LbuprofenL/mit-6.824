#.bash_profile
# Get the aliases and functions
if [ -f ~/.bashrc ]; then
	. ~/.bashrc
fi
# User specific environment and startup programs
GOROOT=/usr/lib/go
GOPATH=/go
PATH=$PATH:/usr/lib/go/bin
export PATH
