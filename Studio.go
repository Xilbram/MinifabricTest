package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type Material struct {
	Descricao  string `json:"descricao"`
	Quantidade int    `json:"quantidade"`
	Owner      string `json:"owner"`
}

type Wand struct {
	ID         string     `json:"id"`
	Materiais  []Material `json:"materiais"`
	Quantidade int        `json:"quantidade"`
	Owner      string     `json:"owner"`
}

// Chaincode Studio para transferencia de materiais e produção de varinhas
type StudioChaincode struct {
}

// Método de inicialização da cadeia.
// stub shim.ChaincodeStubInterface é uma interface que serve como ponte de comunicação entre a CC e a fabric peer (https://ibm.github.io/hlf-internals/shim-architecture/components/chaincode-stub-interface/)]
// Um contador de
func (cc *StudioChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	materiais := []Material{
		{Descricao: "Cabos de ébano", Quantidade: 100, Owner: "Ezequiel"},
		{Descricao: "Rubis", Quantidade: 50, Owner: "Salomão"},
	}

	//Itera e coloca os materiais na ledger. Devem ser buscados via descrição (string)
	for _, material := range materiais {
		materialBytes, err := json.Marshal(material)
		if err != nil {
			return shim.Error(fmt.Sprintf("Erro ao serializar o material %s: %s", material.Descricao, err.Error()))
		}

		err = stub.PutState(material.Descricao, materialBytes)
		if err != nil {
			return shim.Error(fmt.Sprintf("Erro ao salvar o material %s: %s", material.Descricao, err.Error()))
		}
	}

	fmt.Println("O ledger foi inicializado")
	return shim.Success(nil)
}

// Possiveis funções: QueryMaterials, QueryWands, ExchangeMaterials, CreateWand
func (t *StudioChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Invoke é chamado")
	function, args := stub.GetFunctionAndParameters()
	if function == "QueryMaterials" {
		// Lista os materiais disponiveis na rede
		return t.QueryMaterials(stub, args)
	} //else if function == "delete" {
	// Deletes an entity from its state
	//return t.delete(stub, args)
	//} //else if function == "query" {
	// the old "Query" is now implemtned in invoke
	//return t.query(stub, args)
	//}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

// Requer Material.descrição para acessar o material e suas quantidades na rede
func (t *StudioChaincode) QueryMaterials(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Query materials foi chamado")

	var DescMaterial string
	var err error

	if len(args) != 1 {
		return shim.Error("Número incorreto de argumentos. Espera-se que a descrição do material seja inserida")
	}

	DescMaterial = args[0]

	MaterialBytes, err := stub.GetState(DescMaterial)
	if err != nil {
		jsonResp := "{\"Error\":\"Falha ao obter " + DescMaterial + "\"}"
		return shim.Error(jsonResp)
	}
	if MaterialBytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + DescMaterial + "\"}"
		return shim.Error(jsonResp)
	}

	var material Material
	resp := json.Unmarshal(MaterialBytes, &material)
	fmt.Printf("Query Response:%s\n", resp)
	return shim.Success(MaterialBytes)

}

func main() {
	err := shim.Start(new(StudioChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
