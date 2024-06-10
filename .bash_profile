#.bash_profile
# Get the aliases and functions
if [ -f ~/.bashrc ]; then
	. ~/.bashrc
fi
# User specific environment and startup programs
GOROOT=/usr/lib/go
GOPATH=/home
PATH=$PATH:$HOME/bin:$GOROOT/bin:$GOPATH/bin
export PATH