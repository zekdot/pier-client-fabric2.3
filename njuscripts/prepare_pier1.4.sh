# It's where you work, you should also put this script here.
export WORKSPACE=$HOME/bitxhub/
# BitXHub's official pier-client project cloned from GitHub.
export PIER_CLIENT_PATH=${WORKSPACE}/pier-client-fabric
# BitXHub's official relay-chain project cloned from GitHub.
export BITXHUB_PATH=$WORKSPACE/bitxhub
# This two parameter should not be modified if you don't know how to use them.
export CRYPTO_CONFIG_PATH=$WORKSPACE/crypto-config
export FABRIC_RULE_PATH=$WORKSPACE/fabric_rule.wasm

function compile_plugins() {
    cd $PIER_CLIENT_PATH
    git checkout v1.0.0-rc1
    make fabric1.4
}

function prepare_pier() {
    rm -rf $HOME/.pier
    pier --repo=$HOME/.pier init
    # replace line 19~22 to four addresses of bitxhub/scripts/build/genesis.json
    head -n 18 $HOME/.pier/pier.toml > $HOME/.pier/pier.toml.new
    head -n 6 $BITXHUB_PATH/scripts/build_solo/genesis.json | tail -n -4 >> $HOME/.pier/pier.toml.new
    tail -n 7 $HOME/.pier/pier.toml >> $HOME/.pier/pier.toml.new
    # cat pier.toml.new | sed 
    export TEMP=`head -n 6 $BITXHUB_PATH/scripts/build_solo/genesis.json | tail -n -1`
    cat $HOME/.pier/pier.toml.new | sed '22c \ \ \ \ '"$TEMP"',' > $HOME/.pier/pier.toml
    # mv $HOME/.pier/pier.toml.new $HOME/.pier/pier.toml
    # 创建插件文件夹并进行拷贝
    mkdir $HOME/.pier/plugins
    cp $PIER_CLIENT_PATH/build/fabric-client-1.4.so $HOME/.pier/plugins/
    cp -r $PIER_CLIENT_PATH/config $HOME/.pier/fabric
    # 准备加密材料
    cp -r $CRYPTO_CONFIG_PATH $HOME/.pier/fabric/
    # ls $HOME/.pier/fabric/
    # 复制Fabric上验证人证书
    cp $HOME/.pier/fabric/crypto-config/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/msp/signcerts/peer1.org2.example.com-cert.pem $HOME/.pier/fabric/fabric.validators
    
    # 修改网络配置和路径
    sed -i 's:\${CONFIG_PATH}:\${HOME}/.pier/fabric:g' $HOME/.pier/fabric/config.yaml
    sed -i 's/host.docker.internal/localhost/g' $HOME/.pier/fabric/config.yaml
    # 可能需要修改$HOME/.pier/fabric/fabric.toml，这里只替换一下第一行的内容就行
    cat $PIER_CLIENT_PATH/config/fabric.toml | sed '1c addr = "localhost:7053"' > $HOME/.pier/fabric/fabric.toml
    # 最后对中继链进行注册
    pier --repo $HOME/.pier appchain register \
        --name hf14 \
        --type fabric \
        --desc chainA-description \
        --version 1.4.7 \
        --validators $HOME/.pier/fabric/fabric.validators
    # 部署验证规则
    pier rule deploy --path $FABRIC_RULE_PATH
    export PIER_ID=`pier --repo=$HOME/.pier id`
    echo "pier is ready, address is $PIER_ID, you need to save this value for send cross-chain request"
    echo "run following code to start pier"
    echo "pier --repo=$HOME/.pier start"
}
function start() {
    compile_plugins
    prepare_pier
}
start