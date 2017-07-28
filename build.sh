if [ -f ./bin/main ];then
mv ./bin/main ./bin/main.bak
fi
GITLOGVERSION=`git log --oneline --decorate --graph | sed -n 1p | awk '{print $2}'` 
echo "start build $GITLOGVERSION"
gb build -ldflags "-X main.GitLog=$GITLOGVERSION"