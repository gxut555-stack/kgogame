#!/bin/bash

# 检查环境变量 PUBLIC_PROJECT_ENV 是否定义
if [ -z "${PUBLIC_PROJECT_ENV+x}" ]; then
    echo "环境变量 PUBLIC_PROJECT_ENV 未定义，需要在.bash_profile中定义"
    exit 1
fi

# 检查 PUBLIC_PROJECT_ENV 对应的文件是否存在
if [ ! -f "$PUBLIC_PROJECT_ENV" ]; then
    echo "全局配置文件 '$PUBLIC_PROJECT_ENV' 不存在"
    exit 1
fi

# 如果文件存在，则引用（source）该文件
. "${PUBLIC_PROJECT_ENV}"

cd $(dirname $0)

#检测参数
CheckCommonArgNum $# 2 "Usage: $0 build|run|buildRun local|dev|test"

ENV_NAME=$2
PLOG_CFG="etc/logConf.$ENV_NAME.ini"

#检测配置文件
CheckFile $PLOG_CFG

PROCESS_NAME="frontProxy"

#执行编译
DoCommonBuildOrRun $1 $PROCESS_NAME "-tags myzk *.go"

killall $PROCESS_NAME
rm -f *nohup*

nohup ./$PROCESS_NAME -local :5002 -remote 172.17.0.1:5001 -plogCfg "$PLOG_CFG"  >stdout.1.txt 2>stderr.1.txt &
nohup ./$PROCESS_NAME -local :5003 -remote 172.17.0.1:29001 -plogCfg "$PLOG_CFG"  >stdout.1.txt 2>stderr.1.txt &
#nohup ./$PROCESS_NAME -acPort 29001 -v 0.0.0.1 -servId 2 -name AccessService -h 127.0.0.1:8002 -agollo $AGOLLO_CFG -plogCfg "$PLOG_CFG" >stdout.2.txt 2>stderr.2.txt &

ps -ef | grep $PROCESS_NAME | grep  -v grep
echo "======================"
