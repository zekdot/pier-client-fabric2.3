# It's where you work, you should also put this script here.
export WORKSPACE=$HOME/bitxhub
# It's our project's path
export PIER_CLIENT_PATH=$WORKSPACE/pier-client-fabric2.3
# It's fabric-samples's path. fabric-samples is cloned from GitHub.
export FABRIC_SAMPLE_PATH=$WORKSPACE/fabric-samples
# bitxhub url that deploy relay-chain
export REMOTE_BITXHUB_PATH=xy@xy:/home/xy/bitxhub/bitxhub
# relay-chain's server ip
export REMOTE_IP="172.19.241.113"

function compile_plugins() {
    cd $PIER_CLIENT_PATH
    make fabric2.3
}

function prepare_pier() {
    rm -rf $HOME/.pier
    pier --repo=$HOME/.pier init
    # copy address from relay-chain server and modify pier config file
    scp $REMOTE_BITXHUB_PATH/scripts/build_solo/genesis.json $WORKSPACE
    head -n 18 $HOME/.pier/pier.toml > $HOME/.pier/pier.toml.new
    head -n 6 $WORKSPACE/genesis.json | tail -n -4 >> $HOME/.pier/pier.toml.new
    tail -n 7 $HOME/.pier/pier.toml >> $HOME/.pier/pier.toml.new

    # cat pier.toml.new | sed 
    export TEMP=`head -n 6 $WORKSPACE/genesis.json | tail -n -1`
    echo $TEMP
    cat $HOME/.pier/pier.toml.new | sed '22c \ \ \ \ '"$TEMP"',' > $HOME/.pier/pier.toml
    sed -i '16c addr = '\"$REMOTE_IP:60011\"'' $HOME/.pier/pier.toml
    sed -i '28c plugin = "fabric2.3.so"' $HOME/.pier/pier.toml
    sed -i '29c config = "fabric2.3"' $HOME/.pier/pier.toml
    rm $HOME/.pier/pier.toml.new
    # create plugin dir and copy necessary file
    mkdir $HOME/.pier/plugins
    cp $PIER_CLIENT_PATH/build/fabric2.3.so $HOME/.pier/plugins/
    mkdir $HOME/.pier/fabric2.3
    cp -r $PIER_CLIENT_PATH/config/* $HOME/.pier/fabric2.3
    cp -r $FABRIC_SAMPLE_PATH/test-network/organizations $HOME/.pier/fabric2.3/
    cp $HOME/.pier/fabric2.3/organizations/peerOrganizations/org1.example.com/msp/tlscacerts/tlsca.org1.example.com-cert.pem $HOME/.pier/fabric2.3/fabric.validators
    
    # register to relay-chain
    pier --repo $HOME/.pier appchain register \
        --name fabric2.3 \
        --type fabric2.3 \
        --desc chainC-description \
        --version 1.0.0 \
        --validators $HOME/.pier/fabric2.3/fabric.validators
    pier rule deploy --path $WORKSPACE/fabric_rule.wasm
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