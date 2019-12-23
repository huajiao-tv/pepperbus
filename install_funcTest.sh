OLD_GOPATH=$GOPATH
CURDIR=`pwd`
OLD_GOBIN=$GOBIN
export GOPATH=$CURDIR"/../../../.."
export GOBIN=$CURDIR"/bin/"
echo $GOPATH
echo $GOBIN
mkdir -p $GOBIN
go build -o $GOBIN/pepperbus
go install github.com/huajiao-tv/pepperbus/tester/funcTest
