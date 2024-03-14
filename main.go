package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

//Define o holder dos objetos Material e Wand. ID é unico
type Owner struct{
	ObjectType string `json:"docType"`
	Materiais  []Material `json:"materiais"`
	Wands []Wand `json:"wands"`
	Id string `json:"id"`
}

//Objeto generico representante de matéria prima. Atrelado a 1 owner 
type Material struct {
	ObjectType string `json:"docType"`
	Descricao  string `json:"descricao"`
	Quantidade int    `json:"quantidade"`
	Owner      string `json:"owner"`
}

//Objeto refinado a partir de pelo menos 2 matérias primas. Atrelado a 1 owner
type Wand struct {
	ObjectType string `json:"docType"`
	Materiais  []Material `json:"materiais"`
	Quantidade int        `json:"quantidade"`
	Owner      string     `json:"owner"`
}


// Chaincode Studio para transferencia de materiais e produção de varinhas
type StudioChaincode struct {
}

// Método de inicialização da cadeia.
func (cc *StudioChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *StudioChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Invoke é chamado")
	function, args := stub.GetFunctionAndParameters()
	if function == "getMaterials" {
		// Lista os materiais disponíveis na rede
		return t.QueryMateriais(stub)
	} else if function == "getWands" {
		// Lista todas as varinhas produzidas na rede
		return t.QueryWands(stub)
	} else if function == "initOwner" {
		// Produz um novo owner
		return t.initOwner(stub, args)
	}else if function == "swapMaterials" {
		// Troca materiais entre organizações
		return t.TransferirMateriais(stub, args)
	} else if function == "QueryOwner" {
		// Pega dados do owner
		return t.QueryOwner(stub, args)
	}else if function == "createWand" {
		// Produz uma nova varinha
		return t.CreateSingleWand(stub, args)
	}else if function == "initMaterial" {
		// Cria um novo material associado a um owner
		return t.initMaterial(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"getMaterials\",\"initOwner\",\"QueryOwner\" ,\"initMaterial\", \"getWands\", \"swapMaterials\", or \"createWand\"")
}

//Cria um novo material na ledger
//Possui como entrada a descrição do material, sua quantidade e o ID do seu owner
func (cc *StudioChaincode) initMaterial(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Numero incorreto de argumentos. Espera-se 3: descricao do material, quantidade e ID do dono")
	}

	descricao := args[0]
	quantity, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Quantidade deve ser um numero inteiro")
	}
	ownerID := args[2]

	// Pega o owner da ledger
	ownerAsBytes, err := stub.GetState(ownerID)
	if err != nil {
		return shim.Error("Failed to get ownerId " + err.Error())
	}
	var owner Owner
	err = json.Unmarshal(ownerAsBytes, &owner)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to deserialize owner: %s", err.Error()))
	}

	material := Material{
		ObjectType: "material",
		Descricao:  descricao,
		Quantidade: quantity,
		Owner:      ownerID,
	}

	// Adiciona o novo material ao slice de materias do owner
	owner.Materiais = append(owner.Materiais, material)

	// Serializa o owner com os materiais atualizados
	ownerBytes, err := json.Marshal(owner)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to serialize the owner: %s", err.Error()))
	}

	// Coloca o owner atualizado de volta na ledger
	err = stub.PutState(ownerID, ownerBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to save the owner: %s", err.Error()))
	}

	return shim.Success(nil)
}

//Init owner inicializa um novo Owner na ledger. Deve usar um ID único
//Possui como entrada um ID(string)
func (cc *StudioChaincode) initOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	if len(args) != 1 {
		return shim.Error("Número incorreto de argumentos. Espera-se 1: ID do dono")
	}

	ownerID := args[0]
	ownerAsBytes, err := stub.GetState(ownerID)
	if err != nil {
		return shim.Error("Failed to get ownerId " + err.Error())
	}
	if ownerAsBytes != nil{
		return shim.Error("This owner already exists: " + ownerID)
	}

	var materials []Material
	var wands []Wand

	owner := Owner{
		ObjectType: "owner",
		Materiais: materials,
		Wands: wands,
		Id: ownerID,
	}

	ownerBytes, err := json.Marshal(owner)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to serialize the owner: %s", err.Error()))
	}
	err = stub.PutState(ownerID, ownerBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to save the owner: %s", err.Error()))
	}

	return shim.Success(nil)
}

//Query owner pega todas dados de um owner na ledger. 
//Tem como entrada o ID do owner
func (t *StudioChaincode) QueryOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Número incorreto de argumentos. Espera-se 1: ID do dono")
	}

	ownerID := args[0]
	ownerBytes, err := stub.GetState(ownerID)
    if err != nil {
		return shim.Error(fmt.Sprintf("Error retrieving owner %s", err.Error()))
	}

	return shim.Success(ownerBytes)
}

//Query materials pega todas wands disponiveis na ledger. 
//Não possui parametros de entrada
func (t *StudioChaincode) QueryWands(stub shim.ChaincodeStubInterface) pb.Response {
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao obter varinhas: %s", err.Error()))
	}
	defer resultsIterator.Close()

	var wands []Wand

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("Erro ao iterar sobre varinhas: %s", err.Error()))
		}
		var myOwner Owner
		err = json.Unmarshal(queryResponse.Value, &myOwner)
		if err != nil {
			return shim.Error(fmt.Sprintf("Erro ao deserializar varinha: %s", err.Error()))
		}

		for _, eachWand := range myOwner.Wands {
			wands = append(wands, eachWand)
		}
	}

	wandsBytes, err := json.Marshal(wands)
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao serializar varinhas: %s", err.Error()))
	}

	return shim.Success(wandsBytes)
}

//Query materials pega todos materias disponiveis na ledger. 
//Não possui parametros de entrada
func (t *StudioChaincode) QueryMateriais(stub shim.ChaincodeStubInterface) pb.Response {
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao obter materiais: %s", err.Error()))
	}
	defer resultsIterator.Close()

	var materials []Material


	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("Erro ao iterar sobre materiais: %s", err.Error()))
		}
		var myOwner Owner
		err = json.Unmarshal(queryResponse.Value, &myOwner)
		if err != nil {
			return shim.Error(fmt.Sprintf("Erro ao deserializar materiais: %s", err.Error()))
		}

		for _, eachMaterial := range myOwner.Materiais {
			materials = append(materials, eachMaterial)
		}
	}

	materialsBytes, err := json.Marshal(materials)
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao serializar materiais: %s", err.Error()))
	}

	return shim.Success(materialsBytes)
}



//Gera uma nova varinha e registra ela a um owner
//O método pede o id de um owner, verifica os materiais associados ao ID dele e combina 2 materiais diferentes
//Tem como entrada o ID do owner
//Consome todos materiais para criar uma varinha 
//(poderia também haver uma iteração para que a varinha consumisse NxM materiais,
//mas não tinha certeza da lógica a implementar já que é um objeto imaginário)
func (t *StudioChaincode) CreateSingleWand(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Número incorreto de argumentos. Esperando 1: ID do proprietário")
	}

	ownerID := args[0]

	// Pega o owner da ledger
	ownerAsBytes, err := stub.GetState(ownerID)
	if err != nil {
		return shim.Error("Falha ao obter owner " + err.Error())
	}
	if ownerAsBytes == nil {
		return shim.Error("Owner não existe: " + ownerID)
	}
	var owner Owner
	err = json.Unmarshal(ownerAsBytes, &owner)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to deserialize owner: %s", err.Error()))
	}

	// Verifica se o owner tem pelo menos 2 tipos de materiais
	if len(owner.Materiais) < 2 {
		return shim.Error("Owner does not have enough materials to create a wand")
	}

	// Create a new wand with the first two materials
	newWand := Wand{
		ObjectType: "wand",
		Materiais:  owner.Materiais[:2],
		Quantidade: 1,
		Owner:      ownerID,
	}

	// Remove the used materials from the owner's materials
	owner.Materiais = owner.Materiais[2:]

	// Add the new wand to the owner's wands
	owner.Wands = append(owner.Wands, newWand)

	// Serialize the owner with the updated wands and materials
	ownerBytes, err := json.Marshal(owner)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to serialize the owner: %s", err.Error()))
	}

	// Put the updated owner back to the ledger
	err = stub.PutState(ownerID, ownerBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to save the owner: %s", err.Error()))
	}

	return shim.Success(nil)
}


// TransferirMateriais permite a troca de materiais entre 2 orgs
//Possui como entrada de argumentos: Id do enviador, descrição do material a ser enviado, quantidade e ID do recipiente
func (cc *StudioChaincode) TransferirMateriais(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4: sender ID, material description, quantity, receiver ID")
	}

	senderID := args[0]
	materialDescription := args[1]
	quantity, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Quantity must be a valid integer")
	}
	receiverID := args[3]

	// Pega os dados do sender do ledger
	senderBytes, err := stub.GetState(senderID)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get sender owner: %s", err.Error()))
	}
	if senderBytes == nil {
		return shim.Error(fmt.Sprintf("Sender owner not found: %s", senderID))
	}
	var sender Owner
	err = json.Unmarshal(senderBytes, &sender)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to deserialize sender owner: %s", err.Error()))
	}

	// Pega os dados do recipiente do ledger
	receiverBytes, err := stub.GetState(receiverID)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get receiver owner: %s", err.Error()))
	}
	if receiverBytes == nil {
		return shim.Error(fmt.Sprintf("Receiver owner not found: %s", receiverID))
	}
	var receiver Owner
	err = json.Unmarshal(receiverBytes, &receiver)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to deserialize receiver owner: %s", err.Error()))
	}

	// Pega o material especificado dentro do slice do sender
	var foundMaterial *Material
	for i, material := range sender.Materiais {
		if material.Descricao == materialDescription {
			foundMaterial = &sender.Materiais[i]
			break
		}
	}
	if foundMaterial == nil {
		return shim.Error(fmt.Sprintf("Material %s not found in sender's materials", materialDescription))
	}

	// Verifica quantidade de materiais
	if foundMaterial.Quantidade < quantity {
		return shim.Error(fmt.Sprintf("Insufficient quantity of material %s owned by sender %s", materialDescription, senderID))
	}
	foundMaterial.Quantidade -= quantity

	var receiverMaterial *Material
	for i, material := range receiver.Materiais {
		if material.Descricao == materialDescription {
			receiverMaterial = &receiver.Materiais[i]
			break
		}
	}

	// Se o material ja existe no recipiente aumenta sua quantidade
	// Se não existe, adiciona o material ao slice
	if receiverMaterial != nil {
		receiverMaterial.Quantidade += quantity
	} else {
		receiver.Materiais = append(receiver.Materiais, Material{
			ObjectType: "material",
			Descricao:  materialDescription,
			Quantidade: quantity,
			Owner:      receiverID,
		})
	}

	// Serializa e salva o sender
	senderBytes, err = json.Marshal(sender)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to serialize sender owner: %s", err.Error()))
	}
	err = stub.PutState(senderID, senderBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to save updated sender owner: %s", err.Error()))
	}

	// Serializa e salva o recipiente
	receiverBytes, err = json.Marshal(receiver)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to serialize receiver owner: %s", err.Error()))
	}
	err = stub.PutState(receiverID, receiverBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to save updated receiver owner: %s", err.Error()))
	}

	return shim.Success(nil)
}



func main() {
	err := shim.Start(new(StudioChaincode))
	if err != nil {
		fmt.Printf("Error starting Studio chaincode: %s", err)
	}
}

