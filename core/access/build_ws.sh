#!/bin/bash

if [ $# != 2 ]; then
    echo "Usage: $0 build|run|buildRun local|dev|test"
	exit
fi

#获得以太网物理网卡地址
EthernetIP() {
    machine_physics_net=$(ls /sys/class/net/ | grep -v "`ls /sys/devices/virtual/net/`");
    #先过滤网卡，在查找IP，不要再awk中过滤网卡
    local_ip=$(ip addr | grep $machine_physics_net | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}' | head -1);
    localIP=$local_ip
}

cd $(dirname $0)

ENV_NAME=$2
AGOLLO_CFG=${AGOLLO_CFG-"tcp://47.96.182.166:8060/kgogame"}
CFG="etc/cfg.$ENV_NAME.toml"
localIP="127.0.0.1"
EthernetIP

PROCESS_NAME="AccessService"
if [ $1 = "buildRun" ] || [ $1 = "build" ]; then
    go build -o $PROCESS_NAME
    if [ $1 = "build" ]; then
        exit
    fi
elif [ $1 != "run" ];then
    echo "[error]param is not run or buildRun"
    exit
fi

if [ ! -f "$CFG" ]; then
	echo "Config file $CFG is not exists"
	exit
fi

#killall $PROCESS_NAME
#rm -f *nohup*
#nohup ./$PROCESS_NAME -acPort 8100 -v 0.0.0.1 -servId 2 -name AccessService -h $localIP:8002 -agollo $AGOLLO_CFG -cfg "$CFG" -ws >stdout.2.txt 2>stderr.2.txt &  #ws 启动
nohup ./$PROCESS_NAME -acPort 8100 -v 0.0.0.1 -servId 2 -name AccessService -h 127.0.0.1:8002 -agollo $AGOLLO_CFG -cfg "$CFG" >stdout.2.txt 2>stderr.2.txt &  #非ws 启动
ps -ef | grep $PROCESS_NAME | grep  -v grep
