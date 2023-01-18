set -e
docker build -t quay.io/wutiarn/idgen-server-go .
docker push quay.io/wutiarn/idgen-server-go
