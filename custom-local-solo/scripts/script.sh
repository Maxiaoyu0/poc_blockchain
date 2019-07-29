#!/bin/bash

echo
echo " ____    _____      _      ____    _____ "
echo "/ ___|  |_   _|    / \    |  _ \  |_   _|"
echo "\___ \    | |     / _ \   | |_) |   | |  "
echo " ___) |   | |    / ___ \  |  _ <    | |  "
echo "|____/    |_|   /_/   \_\ |_| \_\   |_|  "
echo
echo "Build your first network (BYFN) end-to-end test"
echo
CHANNEL_NAME="$1"
DELAY="$2"
LANGUAGE="$3"
TIMEOUT="$4"
VERBOSE="$5"
: ${CHANNEL_NAME:="mychannel"}
: ${DELAY:="3"}
: ${LANGUAGE:="golang"}
: ${TIMEOUT:="10"}
: ${VERBOSE:="false"}
LANGUAGE=`echo "$LANGUAGE" | tr [:upper:] [:lower:]`
COUNTER=1
MAX_RETRY=10

CC_SRC_PATH="github.com/chaincode/firstTEST/"
if [ "$LANGUAGE" = "node" ]; then
	echo "node not supported yet"
fi

if [ "$LANGUAGE" = "java" ]; then
	echo "java not supported yet"
fi

echo "Channel name : "$CHANNEL_NAME

# import utils
. scripts/utils.sh

createChannel() {
	setGlobals 0 1

	if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		set -x
		peer channel create -o orderer.ticketChain.sfeir.lu:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx >&log.txt
		res=$?
		set +x
	else
		set -x
		peer channel create -o orderer.ticketChain.sfeir.lu:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
		res=$?
		set +x
	fi
	cat log.txt
	verifyResult $res "Channel creation failed"
	echo "===================== Channel '$CHANNEL_NAME' created ===================== "
	echo
}

joinChannel () {
	for org in {1..3}; do
	    for peer in 0 1; do
		joinChannelWithRetry $peer $org
		echo "===================== peer${peer}.$(convertToOrgName $org) joined channel '$CHANNEL_NAME' ===================== "
		sleep $DELAY
		echo
	    done
	done
}

## Create channel
echo "Creating channel..."
createChannel

## Join all the peers to the channel
echo "Having all peers join the channel..."
joinChannel

## Set the anchor peers for each org in the channel
echo "Updating anchor peers for sfeir..."
updateAnchorPeers 0 1
echo "Updating anchor peers for ticketMaster..."
updateAnchorPeers 0 2
echo "Updating anchor peers for bankRoute..."
updateAnchorPeers 0 3

## Install chaincode on peer0.sfeir and peer0.ticketMaster
echo "Installing chaincode on peer0.sfeir..."
installChaincode 0 1
echo "Install chaincode on peer0.ticketMaster..."
installChaincode 0 2
echo "Install chaincode on peer0.bankRoute..."
installChaincode 0 3

# Instantiate chaincode on peer0.ticketMaster
echo "Instantiating chaincode on peer0.ticketMaster..."
instantiateChaincode 0 2

# Query chaincode on peer0.sfeir
# echo "Querying chaincode on peer0.sfeir..."
# chaincodeQuery 0 1 100

sleep 5

# Invoke chaincode on peer0.sfeir and peer0.ticketMaster
echo "Sending invoke transaction on peer0.sfeir peer0.ticketMaster peer0.bankRoute"
chaincodeInvoke 0 1 0 2 0 3

# Install chaincode on peer1.ticketMaster
# echo "Installing chaincode on peer1.ticketMaster..."
# installChaincode 1 2

# Query on chaincode on peer1.ticketMaster, check if the result is 90
# echo "Querying chaincode on peer1.ticketMaster..."
# chaincodeQuery 1 2 90

echo
echo "========= All GOOD, BYFN execution completed =========== "
echo

echo
echo " _____   _   _   ____   "
echo "| ____| | \ | | |  _ \  "
echo "|  _|   |  \| | | | | | "
echo "| |___  | |\  | | |_| | "
echo "|_____| |_| \_| |____/  "
echo

exit 0