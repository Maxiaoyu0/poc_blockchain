package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

/********** MODELS **********/

// RestaurantTicket represents the ticket used to buy some goods
type RestaurantTicket struct {
	serialNo        uint64
	value           uint8
	manufacturer    Organization
	owner           User
	utilisationDate *time.Time
	expirationDate  time.Time
}

// Organization represents an org inside the network (blockchain)
type Organization struct {
	identity string
	name     string
	country  string
	address  string
}

// User represents one of the user able to purchase ticket and use them
type User struct {
	identity string
	name     string
	org      Organization
}

/********** CHAINCODE **********/

// RestaurantTicketCC implements a chaincode to manage a restaurant ticket
type RestaurantTicketCC struct {
}

// Init is called during chaincode instantiation to initialize any data.
func (t *RestaurantTicketCC) Init(stub shim.ChaincodeStubInterface) peer.Response {
	args := stub.GetStringArgs()
	if len(args) < 2 {
		return shim.Error("Incorrect arguments. Expecting at least a key and a value")
	}

	// Set up any variables or restaurant tickets here by calling stub.PutState()
	orgs := [3]Organization{
		{
			identity: "1",
			name:     "Sfeir",
			country:  "Luxembourg",
			address:  "rue Drosbach",
		},
		{
			identity: "2",
			name:     "TicketMaster",
			country:  "Luxembourg",
			address:  "rue differente",
		},
		{
			identity: "3",
			name:     "BankRoute",
			country:  "Luxembourg",
			address:  "rue de la banque",
		},
	}

	users := [3]User{
		{
			identity: "10",
			name:     "Xiaoyu Ma",
			org:      orgs[0],
		},
		{
			identity: "20",
			name:     "Jean Dupont",
			org:      orgs[1],
		},
		{
			identity: "30",
			name:     "Arthur Dujambon",
			org:      orgs[2],
		},
	}

	for i := 0; i < len(orgs); i++ {
		orgAsJSONByte, err := json.Marshal(orgs[i])
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to marshal Organization object as JSON bytes"))
		}
		err = stub.PutState(orgs[i].identity, orgAsJSONByte)
	}

	for i := 0; i < len(users); i++ {
		userAsJSONByte, err := json.Marshal(users[i])
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to marshal User object as JSON bytes"))
		}
		err = stub.PutState(users[i].identity, userAsJSONByte)
	}

	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the restaurant ticket created by Init function. The 'set'
// method may create a new restaurant ticket by specifying a new key-value pair.
func (t *RestaurantTicketCC) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	var result []byte
	var err error
	if fn == "create" {
		result, err = create(stub, args)
	} else if fn == "use" {
		result, err = use(stub, args)
	} else if fn == "convert" {
		result, err = convertToMoney(stub, args)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success(result)
}

// Set stores the restaurant ticket (both key and value) on the ledger.
// args:
//	0: serialNo of the ticket
//	1: value of the ticket
//	2: creator identity
//	3: owner identity
//	4: number of day the ticket will be valid
func create(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	fmt.Println("START create chaincode")
	if len(args) < 5 {
		return nil, fmt.Errorf("Incorrect arguments. Expecting six args: serialNo value manufacturer owner utilisationDate expirationDate")
	}

	var err error

	// args[0] serialNo (uint64)
	serialNoConv, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Incorrect argument. The first argument should be a uint64")
	}
	// args[1] value (uint8)
	valueConv, err := strconv.ParseUint(args[1], 10, 8)
	if err != nil {
		return nil, fmt.Errorf("Incorrect argument. The second argument should be a uint8")
	}

	// args[2] creator (User.identity)
	creatorAsBytes, err := stub.GetState(args[2])
	if err != nil {
		return nil, fmt.Errorf("Invalid argument. The third argument should be the identity of the Creator")
	}
	if creatorAsBytes == nil {
		return nil, fmt.Errorf("Invalid argument. The creator with the identity %s doesn't exist", args[2])
	}
	creator := User{}
	err = json.Unmarshal(creatorAsBytes, &creator)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal the User creator")
	}

	manufacturer := creator.org

	// args[3] owner (User.identity)
	ownerAsBytes, err := stub.GetState(args[3])
	if err != nil {
		return nil, fmt.Errorf("Impossible to get the state of the third argument")
	}
	owner := User{}
	err = json.Unmarshal(ownerAsBytes, &owner)
	if err != nil {
		return nil, fmt.Errorf("Incorrect arguments. Impossible to unmarshal the fourth argument")
	}

	// args[4] expirationDate
	dureeValidite, err := strconv.Atoi(args[4])
	if err != nil {
		return nil, fmt.Errorf("Incorrect argument. The fourth argument, the expiration date should be an integer")
	}

	newTicket := RestaurantTicket{
		serialNo:        uint64(serialNoConv),
		value:           uint8(valueConv),
		manufacturer:    manufacturer,
		owner:           owner,
		utilisationDate: nil,
		expirationDate:  time.Now().AddDate(0, 0, dureeValidite),
	}

	newTicketAsJSONBytes, err := json.Marshal(newTicket)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal newTicket object as JSON bytes")
	}
	err = stub.PutState(strconv.FormatUint(newTicket.serialNo, 10), newTicketAsJSONBytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to set restaurant ticket: %s", args[0])
	}

	fmt.Println("END CC")

	return newTicketAsJSONBytes, nil
}

// Not finished / tested
func use(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	// args[0] restaurantTicket.serialNo (uint64)
	restaurantTicketAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, fmt.Errorf("Impossible to get the state of the RestaurantTicket: key=%s", args[0])
	}
	restaurantTicket := RestaurantTicket{}
	err = json.Unmarshal(restaurantTicketAsBytes, &restaurantTicket)
	if err != nil {
		return nil, fmt.Errorf("Incorrect arguments. Impossible to unmarshal the RestaurantTicket")
	}

	// Check the expiration date
	now := time.Now()
	if restaurantTicket.expirationDate.Before(now) {
		err = stub.DelState(args[0])
		return nil, fmt.Errorf("Error. The RestaurantTicket have expired and is now deleted")
	}

	// Change the owner
	// args[1] newOwner (User.identity)
	newOwnerAsBytes, err := stub.GetState(args[1])
	if newOwnerAsBytes == nil {
		return nil, fmt.Errorf("Incorrect arguments. The newOwner doesn't exist")
	}
	if err != nil {
		return nil, fmt.Errorf("Impossible to get the state of the new owner: key=%s", args[1])
	}
	newOwnerUser := User{}
	err = json.Unmarshal(newOwnerAsBytes, &newOwnerUser)
	if err != nil {
		return nil, fmt.Errorf("Incorrect arguments. Impossible to unmarshal the newOwner")
	}
	restaurantTicket.owner = newOwnerUser

	// Update de utilisationDate
	restaurantTicket.utilisationDate = &now

	restaurantTicketAsJSONBytes, err := json.Marshal(restaurantTicket)

	return restaurantTicketAsJSONBytes, nil
}

// Not finished / tested
func convertToMoney(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	// args[0] restaurantTicket.serialNo (uint64)
	restaurantTicketAsBytes, err := stub.GetState(args[0])
	if restaurantTicketAsBytes == nil {
		return nil, fmt.Errorf("Incorrect arguments. This Restaurant Ticket doesn't exist")
	}
	if err != nil {
		return nil, fmt.Errorf("Impossible to get the RestaurantTicket: key=%s", args[0])
	}

	err = stub.DelState(args[0])
	if err != nil {
		return nil, fmt.Errorf("Impossible to delete RestaurantTicket: key=%s", args[0])
	}

	result := "RestaurantTicket " + args[0] + " deleted"

	return []byte(result), nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(RestaurantTicketCC)); err != nil {
		fmt.Printf("Error starting RestaurantTicketCC chaincode: %s", err)
	}
}
