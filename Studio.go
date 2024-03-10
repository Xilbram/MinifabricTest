package main

import (
	"encoding/json"
	"fmt"
	"strconv"

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

func (t *StudioChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Invoke é chamado")
	function, args := stub.GetFunctionAndParameters()
	if function == "QueryMaterials" {
		// Lista os materiais disponíveis na rede
		return t.QueryMaterials(stub, args)
	} else if function == "QueryWands" {
		// Lista todas as varinhas produzidas na rede
		return t.QueryWands(stub)
	} else if function == "ExchangeMaterials" {
		// Troca materiais entre organizações
		return t.TransferirMateriais(stub, args)
	} else if function == "CreateWand" {
		// Produz uma nova varinha
		return t.CreateWand(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"QueryMaterials\", \"QueryWands\", \"ExchangeMaterials\", or \"CreateWand\"")
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

// Requer Material.descrição para acessar o material e suas quantidades na rede
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

		var wand Wand
		err = json.Unmarshal(queryResponse.Value, &wand)
		if err != nil {
			return shim.Error(fmt.Sprintf("Erro ao deserializar varinha: %s", err.Error()))
		}

		wands = append(wands, wand)
	}

	wandsBytes, err := json.Marshal(wands)
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao serializar varinhas: %s", err.Error()))
	}

	return shim.Success(wandsBytes)
}

// Cria varinhas se a org for o Sr Olivaras
func (t *StudioChaincode) CreateWand(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Número incorreto de argumentos. Esperando 2: descricao, quantidade")
	}

	descricao := args[0]
	quantidade, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("A quantidade deve ser um número inteiro")
	}

	creator, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao obter o criador da transação: %s", err.Error()))
	}

	var materiais []Material
	for i := 0; i < quantidade; i++ {
		materialBytes, err := stub.GetState(descricao)
		if err != nil {
			return shim.Error(fmt.Sprintf("Erro ao obter o material %s: %s", descricao, err.Error()))
		}

		var material Material
		err = json.Unmarshal(materialBytes, &material)
		if err != nil {
			return shim.Error(fmt.Sprintf("Erro ao deserializar o material %s: %s", descricao, err.Error()))
		}

		if material.Quantidade < quantidade {
			return shim.Error(fmt.Sprintf("Quantidade insuficiente do material %s disponível para produção de varinhas", descricao))
		}

		if material.Owner != string(creator) {
			return shim.Error("Você não possui permissão para utilizar este material")
		}

		material.Quantidade -= quantidade
		materiais = append(materiais, material)
	}

	wand := Wand{
		Materiais:  materiais,
		Quantidade: quantidade,
		Owner:      string(creator),
	}

	wandBytes, err := json.Marshal(wand)
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao serializar a varinha: %s", err.Error()))
	}

	err = stub.PutState(descricao, wandBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao salvar a varinha: %s", err.Error()))
	}

	return shim.Success(nil)
}

// TransferirMateriais transfere um material para a organização do Sr. Olivaras
func (cc *StudioChaincode) TransferirMateriais(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 4 {
		return shim.Error("Número incorreto de argumentos. Esperando 4: descricao, quantidade, destinatario")
	}

	descricao := args[0]
	quantidade, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("A quantidade deve ser um número inteiro")
	}
	destinatario := args[2]

	creator, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao obter o criador da transação: %s", err.Error()))
	}

	materialBytes, err := stub.GetState(descricao)
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao obter o material %s: %s", descricao, err.Error()))
	}

	var material Material
	err = json.Unmarshal(materialBytes, &material)
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao deserializar o material %s: %s", descricao, err.Error()))
	}

	if material.Quantidade < quantidade {
		return shim.Error(fmt.Sprintf("Quantidade insuficiente do material %s disponível para transferência", descricao))
	}

	if material.Owner != string(creator) {
		return shim.Error("Você não possui permissão para transferir este material")
	}

	material.Quantidade -= quantidade

	newMaterialBytes, err := json.Marshal(material)
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao serializar o material %s: %s", descricao, err.Error()))
	}

	err = stub.PutState(descricao, newMaterialBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao atualizar o material %s: %s", descricao, err.Error()))
	}

	// Transferir o material para o destinatário
	recipientMaterialBytes, _ := stub.GetState(descricao)
	var recipientMaterial Material
	json.Unmarshal(recipientMaterialBytes, &recipientMaterial)
	recipientMaterial.Quantidade += quantidade
	recipientMaterial.Owner = destinatario

	recipientMaterialBytes, err = json.Marshal(recipientMaterial)
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao serializar o material %s: %s", descricao, err.Error()))
	}

	err = stub.PutState(descricao, recipientMaterialBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao transferir o material %s para %s: %s", descricao, destinatario, err.Error()))
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(StudioChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
