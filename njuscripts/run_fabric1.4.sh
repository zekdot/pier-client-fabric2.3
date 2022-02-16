# It's where you work, you should also put this script here.
export WORKSPACE=$HOME/bitxhub
# It's fabric-samples's path. fabric-samples is cloned from GitHub.
export FABRIC_SAMPLE_PATH=${WORKSPACE}/fabric-samples
# BitXHub's official project cloned from GitHub.
export PIER_CLIENT_PATH=${WORKSPACE}/pier-client-fabric
# This two parameter should not be modified if you don't know how to use them.
export CONFIG_PATH=$WORKSPACE
export CONFIG_YAML=config.yaml

function start_fabric() {
    docker volume prune -f
    cd ${FABRIC_SAMPLE_PATH}/first-network
    ./byfn.sh generate
    ./byfn.sh up -n
    rm -rf ${WORKSPACE}/crypto-config
    cp -rf ${FABRIC_SAMPLE_PATH}/first-network/crypto-config ${WORKSPACE}/crypto-config
}
function dfs_modify_name()
{
    cd $1
    files=`ls`
    for file in $files;
    do
        # Recursive processing folder
        if [ -d $file ];
        then
            echo "step into $file"
            dfs_modify_name $file
        else
        # whether name is client.xx
            echo "deal with $file"
            if [ ${file%.*} = "client" ];
            then
                echo "copy $file to server.${file#*.}"
                cp $file server.${file#*.}
            fi
            
        fi
    done
    cd ..
}
function deploy_contract() {
    cd $WORKSPACE
    dfs_modify_name crypto-config
    # 把跨链合约移动到当前目录下
    cp -rf $PIER_CLIENT_PATH/example/contracts .
    # 部署与实例化
    fabric-cli chaincode install --gopath ./contracts --ccp broker --ccid broker --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel        # broker链码安装
    fabric-cli chaincode instantiate --ccp broker --ccid broker --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel # broker链码实例化
    fabric-cli chaincode install --gopath ./contracts --ccp data_swapper --ccid data_swapper --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel    # data_swapper链码安装
    fabric-cli chaincode instantiate --ccp data_swapper --ccid data_swapper --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel # data_swapper链码实例化
    # register
    fabric-cli chaincode invoke --cid mychannel --ccid=data_swapper \
    --args='{"Func":"register"}' --user Admin --orgid org2 --payload --config "config.yaml"
    # audit
    fabric-cli chaincode invoke --cid mychannel --ccid=broker \
    --args='{"Func":"audit", "Args":["mychannel", "data_swapper", "1"]}' \
    --user Admin --orgid org2 --payload --config "config.yaml"
}
function start() {
    start_fabric
    deploy_contract
}
start
docker ps