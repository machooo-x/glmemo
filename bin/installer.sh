#!/bin/bash

servers=(
    "139.9.133.162"
)

for server in ${servers[@]}; do
    echo "开始更新${server}..."
    ssh root@${server} "mv -f /opt/macho/glmemo/linux-amd64-glmemo /opt/macho/glmemo/linux-amd64-glmemo-back && exit"
    scp ${PWD}/linux-amd64-glmemo root@${server}:/opt/macho/glmemo/linux-amd64-glmemo
    ssh root@${server} "cd /opt/macho/glmemo && chmod +x linux-amd64-glmemo && ./startup.sh 1>/dev/null && exit"
    sleep 1
done