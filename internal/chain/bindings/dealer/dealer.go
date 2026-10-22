// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package dealer

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// TypesRiskParams is an auto generated low-level Go binding around an user-defined struct.
type TypesRiskParams struct {
	InitialMarginRatio   *big.Int
	LiquidationThreshold *big.Int
	LiquidationPriceOff  *big.Int
	InsuranceFeeRate     *big.Int
	MarkPriceSource      common.Address
	Name                 string
	IsRegistered         bool
}

// DealerMetaData contains all meta data concerning the Dealer contract.
var DealerMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_primaryAsset\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"approveTrade\",\"inputs\":[{\"name\":\"orderSender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"tradeData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"},{\"name\":\"\",\"type\":\"int256[]\",\"internalType\":\"int256[]\"},{\"name\":\"\",\"type\":\"int256[]\",\"internalType\":\"int256[]\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"deposit\",\"inputs\":[{\"name\":\"primaryAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"secondaryAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"domainSeparator\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"executeWithdraw\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"fastWithdraw\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"primaryAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"secondaryAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getLiquidationCost\",\"inputs\":[{\"name\":\"perp\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"liquidatedTrader\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"requestPaperAmount\",\"type\":\"int256\",\"internalType\":\"int256\"}],\"outputs\":[{\"name\":\"liqtorPaperChange\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"liqtorCreditChange\",\"type\":\"int256\",\"internalType\":\"int256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getLiquidationPrice\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"perp\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOrderFilledAmount\",\"inputs\":[{\"name\":\"orderHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPositions\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPrimaryCredit\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"int256\",\"internalType\":\"int256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRiskParams\",\"inputs\":[{\"name\":\"perp\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structTypes.RiskParams\",\"components\":[{\"name\":\"initialMarginRatio\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"liquidationThreshold\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"liquidationPriceOff\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"insuranceFeeRate\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"markPriceSource\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"isRegistered\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTraderRisk\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"netValue\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"exposure\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"initialMargin\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maintenanceMargin\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"handleBadDebt\",\"inputs\":[{\"name\":\"liquidatedTrader\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isAllSafe\",\"inputs\":[{\"name\":\"traderList\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isIMSafe\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOperatorValid\",\"inputs\":[{\"name\":\"client\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isSafe\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"openPosition\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"realizePnl\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"pnl\",\"type\":\"int256\",\"internalType\":\"int256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"requestLiquidation\",\"inputs\":[{\"name\":\"executor\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"liquidator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"liquidatedTrader\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"requestPaperAmount\",\"type\":\"int256\",\"internalType\":\"int256\"}],\"outputs\":[{\"name\":\"liqtorPaperChange\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"liqtorCreditChange\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"liqedPaperChange\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"liqedCreditChange\",\"type\":\"int256\",\"internalType\":\"int256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"requestWithdraw\",\"inputs\":[{\"name\":\"primaryAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"secondaryAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setFundingRateKeeper\",\"inputs\":[{\"name\":\"newKeeper\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setInsurance\",\"inputs\":[{\"name\":\"_insurance\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMaxPositionAmount\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"isValid\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOrderSender\",\"inputs\":[{\"name\":\"orderSender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"isValid\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setPerpRiskParams\",\"inputs\":[{\"name\":\"perp\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"params\",\"type\":\"tuple\",\"internalType\":\"structTypes.RiskParams\",\"components\":[{\"name\":\"initialMarginRatio\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"liquidationThreshold\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"liquidationPriceOff\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"insuranceFeeRate\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"markPriceSource\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"isRegistered\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setWithdrawTimeLock\",\"inputs\":[{\"name\":\"timeLock\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"state\",\"inputs\":[],\"outputs\":[{\"name\":\"primaryAsset\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"secondaryAsset\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"withdrawTimeLock\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"fastWithdrawDisabled\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"maxPositionAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"insurance\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"fundingRateKeeper\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateFundingRate\",\"inputs\":[{\"name\":\"perpList\",\"type\":\"address[]\",\"internalType\":\"address[]\"},{\"name\":\"rateList\",\"type\":\"int256[]\",\"internalType\":\"int256[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false}]",
}

// DealerABI is the input ABI used to generate the binding from.
// Deprecated: Use DealerMetaData.ABI instead.
var DealerABI = DealerMetaData.ABI

// Dealer is an auto generated Go binding around an Ethereum contract.
type Dealer struct {
	DealerCaller     // Read-only binding to the contract
	DealerTransactor // Write-only binding to the contract
	DealerFilterer   // Log filterer for contract events
}

// DealerCaller is an auto generated read-only Go binding around an Ethereum contract.
type DealerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DealerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DealerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DealerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DealerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DealerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DealerSession struct {
	Contract     *Dealer           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DealerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DealerCallerSession struct {
	Contract *DealerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// DealerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DealerTransactorSession struct {
	Contract     *DealerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DealerRaw is an auto generated low-level Go binding around an Ethereum contract.
type DealerRaw struct {
	Contract *Dealer // Generic contract binding to access the raw methods on
}

// DealerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DealerCallerRaw struct {
	Contract *DealerCaller // Generic read-only contract binding to access the raw methods on
}

// DealerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DealerTransactorRaw struct {
	Contract *DealerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDealer creates a new instance of Dealer, bound to a specific deployed contract.
func NewDealer(address common.Address, backend bind.ContractBackend) (*Dealer, error) {
	contract, err := bindDealer(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Dealer{DealerCaller: DealerCaller{contract: contract}, DealerTransactor: DealerTransactor{contract: contract}, DealerFilterer: DealerFilterer{contract: contract}}, nil
}

// NewDealerCaller creates a new read-only instance of Dealer, bound to a specific deployed contract.
func NewDealerCaller(address common.Address, caller bind.ContractCaller) (*DealerCaller, error) {
	contract, err := bindDealer(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DealerCaller{contract: contract}, nil
}

// NewDealerTransactor creates a new write-only instance of Dealer, bound to a specific deployed contract.
func NewDealerTransactor(address common.Address, transactor bind.ContractTransactor) (*DealerTransactor, error) {
	contract, err := bindDealer(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DealerTransactor{contract: contract}, nil
}

// NewDealerFilterer creates a new log filterer instance of Dealer, bound to a specific deployed contract.
func NewDealerFilterer(address common.Address, filterer bind.ContractFilterer) (*DealerFilterer, error) {
	contract, err := bindDealer(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DealerFilterer{contract: contract}, nil
}

// bindDealer binds a generic wrapper to an already deployed contract.
func bindDealer(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := DealerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Dealer *DealerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Dealer.Contract.DealerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Dealer *DealerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Dealer.Contract.DealerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Dealer *DealerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Dealer.Contract.DealerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Dealer *DealerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Dealer.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Dealer *DealerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Dealer.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Dealer *DealerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Dealer.Contract.contract.Transact(opts, method, params...)
}

// DomainSeparator is a free data retrieval call binding the contract method 0xf698da25.
//
// Solidity: function domainSeparator() view returns(bytes32)
func (_Dealer *DealerCaller) DomainSeparator(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "domainSeparator")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DomainSeparator is a free data retrieval call binding the contract method 0xf698da25.
//
// Solidity: function domainSeparator() view returns(bytes32)
func (_Dealer *DealerSession) DomainSeparator() ([32]byte, error) {
	return _Dealer.Contract.DomainSeparator(&_Dealer.CallOpts)
}

// DomainSeparator is a free data retrieval call binding the contract method 0xf698da25.
//
// Solidity: function domainSeparator() view returns(bytes32)
func (_Dealer *DealerCallerSession) DomainSeparator() ([32]byte, error) {
	return _Dealer.Contract.DomainSeparator(&_Dealer.CallOpts)
}

// GetLiquidationCost is a free data retrieval call binding the contract method 0x59cd592d.
//
// Solidity: function getLiquidationCost(address perp, address liquidatedTrader, int256 requestPaperAmount) view returns(int256 liqtorPaperChange, int256 liqtorCreditChange)
func (_Dealer *DealerCaller) GetLiquidationCost(opts *bind.CallOpts, perp common.Address, liquidatedTrader common.Address, requestPaperAmount *big.Int) (struct {
	LiqtorPaperChange  *big.Int
	LiqtorCreditChange *big.Int
}, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "getLiquidationCost", perp, liquidatedTrader, requestPaperAmount)

	outstruct := new(struct {
		LiqtorPaperChange  *big.Int
		LiqtorCreditChange *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.LiqtorPaperChange = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.LiqtorCreditChange = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetLiquidationCost is a free data retrieval call binding the contract method 0x59cd592d.
//
// Solidity: function getLiquidationCost(address perp, address liquidatedTrader, int256 requestPaperAmount) view returns(int256 liqtorPaperChange, int256 liqtorCreditChange)
func (_Dealer *DealerSession) GetLiquidationCost(perp common.Address, liquidatedTrader common.Address, requestPaperAmount *big.Int) (struct {
	LiqtorPaperChange  *big.Int
	LiqtorCreditChange *big.Int
}, error) {
	return _Dealer.Contract.GetLiquidationCost(&_Dealer.CallOpts, perp, liquidatedTrader, requestPaperAmount)
}

// GetLiquidationCost is a free data retrieval call binding the contract method 0x59cd592d.
//
// Solidity: function getLiquidationCost(address perp, address liquidatedTrader, int256 requestPaperAmount) view returns(int256 liqtorPaperChange, int256 liqtorCreditChange)
func (_Dealer *DealerCallerSession) GetLiquidationCost(perp common.Address, liquidatedTrader common.Address, requestPaperAmount *big.Int) (struct {
	LiqtorPaperChange  *big.Int
	LiqtorCreditChange *big.Int
}, error) {
	return _Dealer.Contract.GetLiquidationCost(&_Dealer.CallOpts, perp, liquidatedTrader, requestPaperAmount)
}

// GetLiquidationPrice is a free data retrieval call binding the contract method 0x1867b480.
//
// Solidity: function getLiquidationPrice(address trader, address perp) view returns(uint256)
func (_Dealer *DealerCaller) GetLiquidationPrice(opts *bind.CallOpts, trader common.Address, perp common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "getLiquidationPrice", trader, perp)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLiquidationPrice is a free data retrieval call binding the contract method 0x1867b480.
//
// Solidity: function getLiquidationPrice(address trader, address perp) view returns(uint256)
func (_Dealer *DealerSession) GetLiquidationPrice(trader common.Address, perp common.Address) (*big.Int, error) {
	return _Dealer.Contract.GetLiquidationPrice(&_Dealer.CallOpts, trader, perp)
}

// GetLiquidationPrice is a free data retrieval call binding the contract method 0x1867b480.
//
// Solidity: function getLiquidationPrice(address trader, address perp) view returns(uint256)
func (_Dealer *DealerCallerSession) GetLiquidationPrice(trader common.Address, perp common.Address) (*big.Int, error) {
	return _Dealer.Contract.GetLiquidationPrice(&_Dealer.CallOpts, trader, perp)
}

// GetOrderFilledAmount is a free data retrieval call binding the contract method 0x55559a86.
//
// Solidity: function getOrderFilledAmount(bytes32 orderHash) view returns(uint256)
func (_Dealer *DealerCaller) GetOrderFilledAmount(opts *bind.CallOpts, orderHash [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "getOrderFilledAmount", orderHash)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOrderFilledAmount is a free data retrieval call binding the contract method 0x55559a86.
//
// Solidity: function getOrderFilledAmount(bytes32 orderHash) view returns(uint256)
func (_Dealer *DealerSession) GetOrderFilledAmount(orderHash [32]byte) (*big.Int, error) {
	return _Dealer.Contract.GetOrderFilledAmount(&_Dealer.CallOpts, orderHash)
}

// GetOrderFilledAmount is a free data retrieval call binding the contract method 0x55559a86.
//
// Solidity: function getOrderFilledAmount(bytes32 orderHash) view returns(uint256)
func (_Dealer *DealerCallerSession) GetOrderFilledAmount(orderHash [32]byte) (*big.Int, error) {
	return _Dealer.Contract.GetOrderFilledAmount(&_Dealer.CallOpts, orderHash)
}

// GetPositions is a free data retrieval call binding the contract method 0x3eeb530e.
//
// Solidity: function getPositions(address trader) view returns(address[])
func (_Dealer *DealerCaller) GetPositions(opts *bind.CallOpts, trader common.Address) ([]common.Address, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "getPositions", trader)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetPositions is a free data retrieval call binding the contract method 0x3eeb530e.
//
// Solidity: function getPositions(address trader) view returns(address[])
func (_Dealer *DealerSession) GetPositions(trader common.Address) ([]common.Address, error) {
	return _Dealer.Contract.GetPositions(&_Dealer.CallOpts, trader)
}

// GetPositions is a free data retrieval call binding the contract method 0x3eeb530e.
//
// Solidity: function getPositions(address trader) view returns(address[])
func (_Dealer *DealerCallerSession) GetPositions(trader common.Address) ([]common.Address, error) {
	return _Dealer.Contract.GetPositions(&_Dealer.CallOpts, trader)
}

// GetPrimaryCredit is a free data retrieval call binding the contract method 0x8e771c5c.
//
// Solidity: function getPrimaryCredit(address trader) view returns(int256)
func (_Dealer *DealerCaller) GetPrimaryCredit(opts *bind.CallOpts, trader common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "getPrimaryCredit", trader)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPrimaryCredit is a free data retrieval call binding the contract method 0x8e771c5c.
//
// Solidity: function getPrimaryCredit(address trader) view returns(int256)
func (_Dealer *DealerSession) GetPrimaryCredit(trader common.Address) (*big.Int, error) {
	return _Dealer.Contract.GetPrimaryCredit(&_Dealer.CallOpts, trader)
}

// GetPrimaryCredit is a free data retrieval call binding the contract method 0x8e771c5c.
//
// Solidity: function getPrimaryCredit(address trader) view returns(int256)
func (_Dealer *DealerCallerSession) GetPrimaryCredit(trader common.Address) (*big.Int, error) {
	return _Dealer.Contract.GetPrimaryCredit(&_Dealer.CallOpts, trader)
}

// GetRiskParams is a free data retrieval call binding the contract method 0x4f33cb53.
//
// Solidity: function getRiskParams(address perp) view returns((uint256,uint256,uint256,uint256,address,string,bool))
func (_Dealer *DealerCaller) GetRiskParams(opts *bind.CallOpts, perp common.Address) (TypesRiskParams, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "getRiskParams", perp)

	if err != nil {
		return *new(TypesRiskParams), err
	}

	out0 := *abi.ConvertType(out[0], new(TypesRiskParams)).(*TypesRiskParams)

	return out0, err

}

// GetRiskParams is a free data retrieval call binding the contract method 0x4f33cb53.
//
// Solidity: function getRiskParams(address perp) view returns((uint256,uint256,uint256,uint256,address,string,bool))
func (_Dealer *DealerSession) GetRiskParams(perp common.Address) (TypesRiskParams, error) {
	return _Dealer.Contract.GetRiskParams(&_Dealer.CallOpts, perp)
}

// GetRiskParams is a free data retrieval call binding the contract method 0x4f33cb53.
//
// Solidity: function getRiskParams(address perp) view returns((uint256,uint256,uint256,uint256,address,string,bool))
func (_Dealer *DealerCallerSession) GetRiskParams(perp common.Address) (TypesRiskParams, error) {
	return _Dealer.Contract.GetRiskParams(&_Dealer.CallOpts, perp)
}

// GetTraderRisk is a free data retrieval call binding the contract method 0x6e753010.
//
// Solidity: function getTraderRisk(address trader) view returns(int256 netValue, uint256 exposure, uint256 initialMargin, uint256 maintenanceMargin)
func (_Dealer *DealerCaller) GetTraderRisk(opts *bind.CallOpts, trader common.Address) (struct {
	NetValue          *big.Int
	Exposure          *big.Int
	InitialMargin     *big.Int
	MaintenanceMargin *big.Int
}, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "getTraderRisk", trader)

	outstruct := new(struct {
		NetValue          *big.Int
		Exposure          *big.Int
		InitialMargin     *big.Int
		MaintenanceMargin *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.NetValue = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Exposure = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.InitialMargin = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.MaintenanceMargin = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetTraderRisk is a free data retrieval call binding the contract method 0x6e753010.
//
// Solidity: function getTraderRisk(address trader) view returns(int256 netValue, uint256 exposure, uint256 initialMargin, uint256 maintenanceMargin)
func (_Dealer *DealerSession) GetTraderRisk(trader common.Address) (struct {
	NetValue          *big.Int
	Exposure          *big.Int
	InitialMargin     *big.Int
	MaintenanceMargin *big.Int
}, error) {
	return _Dealer.Contract.GetTraderRisk(&_Dealer.CallOpts, trader)
}

// GetTraderRisk is a free data retrieval call binding the contract method 0x6e753010.
//
// Solidity: function getTraderRisk(address trader) view returns(int256 netValue, uint256 exposure, uint256 initialMargin, uint256 maintenanceMargin)
func (_Dealer *DealerCallerSession) GetTraderRisk(trader common.Address) (struct {
	NetValue          *big.Int
	Exposure          *big.Int
	InitialMargin     *big.Int
	MaintenanceMargin *big.Int
}, error) {
	return _Dealer.Contract.GetTraderRisk(&_Dealer.CallOpts, trader)
}

// IsAllSafe is a free data retrieval call binding the contract method 0x181f4922.
//
// Solidity: function isAllSafe(address[] traderList) view returns(bool)
func (_Dealer *DealerCaller) IsAllSafe(opts *bind.CallOpts, traderList []common.Address) (bool, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "isAllSafe", traderList)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAllSafe is a free data retrieval call binding the contract method 0x181f4922.
//
// Solidity: function isAllSafe(address[] traderList) view returns(bool)
func (_Dealer *DealerSession) IsAllSafe(traderList []common.Address) (bool, error) {
	return _Dealer.Contract.IsAllSafe(&_Dealer.CallOpts, traderList)
}

// IsAllSafe is a free data retrieval call binding the contract method 0x181f4922.
//
// Solidity: function isAllSafe(address[] traderList) view returns(bool)
func (_Dealer *DealerCallerSession) IsAllSafe(traderList []common.Address) (bool, error) {
	return _Dealer.Contract.IsAllSafe(&_Dealer.CallOpts, traderList)
}

// IsIMSafe is a free data retrieval call binding the contract method 0xb52b886c.
//
// Solidity: function isIMSafe(address trader) view returns(bool)
func (_Dealer *DealerCaller) IsIMSafe(opts *bind.CallOpts, trader common.Address) (bool, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "isIMSafe", trader)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsIMSafe is a free data retrieval call binding the contract method 0xb52b886c.
//
// Solidity: function isIMSafe(address trader) view returns(bool)
func (_Dealer *DealerSession) IsIMSafe(trader common.Address) (bool, error) {
	return _Dealer.Contract.IsIMSafe(&_Dealer.CallOpts, trader)
}

// IsIMSafe is a free data retrieval call binding the contract method 0xb52b886c.
//
// Solidity: function isIMSafe(address trader) view returns(bool)
func (_Dealer *DealerCallerSession) IsIMSafe(trader common.Address) (bool, error) {
	return _Dealer.Contract.IsIMSafe(&_Dealer.CallOpts, trader)
}

// IsOperatorValid is a free data retrieval call binding the contract method 0xa5fcf803.
//
// Solidity: function isOperatorValid(address client, address operator) view returns(bool)
func (_Dealer *DealerCaller) IsOperatorValid(opts *bind.CallOpts, client common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "isOperatorValid", client, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorValid is a free data retrieval call binding the contract method 0xa5fcf803.
//
// Solidity: function isOperatorValid(address client, address operator) view returns(bool)
func (_Dealer *DealerSession) IsOperatorValid(client common.Address, operator common.Address) (bool, error) {
	return _Dealer.Contract.IsOperatorValid(&_Dealer.CallOpts, client, operator)
}

// IsOperatorValid is a free data retrieval call binding the contract method 0xa5fcf803.
//
// Solidity: function isOperatorValid(address client, address operator) view returns(bool)
func (_Dealer *DealerCallerSession) IsOperatorValid(client common.Address, operator common.Address) (bool, error) {
	return _Dealer.Contract.IsOperatorValid(&_Dealer.CallOpts, client, operator)
}

// IsSafe is a free data retrieval call binding the contract method 0x769a28ac.
//
// Solidity: function isSafe(address trader) view returns(bool)
func (_Dealer *DealerCaller) IsSafe(opts *bind.CallOpts, trader common.Address) (bool, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "isSafe", trader)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsSafe is a free data retrieval call binding the contract method 0x769a28ac.
//
// Solidity: function isSafe(address trader) view returns(bool)
func (_Dealer *DealerSession) IsSafe(trader common.Address) (bool, error) {
	return _Dealer.Contract.IsSafe(&_Dealer.CallOpts, trader)
}

// IsSafe is a free data retrieval call binding the contract method 0x769a28ac.
//
// Solidity: function isSafe(address trader) view returns(bool)
func (_Dealer *DealerCallerSession) IsSafe(trader common.Address) (bool, error) {
	return _Dealer.Contract.IsSafe(&_Dealer.CallOpts, trader)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Dealer *DealerCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Dealer *DealerSession) Owner() (common.Address, error) {
	return _Dealer.Contract.Owner(&_Dealer.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Dealer *DealerCallerSession) Owner() (common.Address, error) {
	return _Dealer.Contract.Owner(&_Dealer.CallOpts)
}

// State is a free data retrieval call binding the contract method 0xc19d93fb.
//
// Solidity: function state() view returns(address primaryAsset, address secondaryAsset, uint256 withdrawTimeLock, bool fastWithdrawDisabled, uint256 maxPositionAmount, address insurance, address fundingRateKeeper)
func (_Dealer *DealerCaller) State(opts *bind.CallOpts) (struct {
	PrimaryAsset         common.Address
	SecondaryAsset       common.Address
	WithdrawTimeLock     *big.Int
	FastWithdrawDisabled bool
	MaxPositionAmount    *big.Int
	Insurance            common.Address
	FundingRateKeeper    common.Address
}, error) {
	var out []interface{}
	err := _Dealer.contract.Call(opts, &out, "state")

	outstruct := new(struct {
		PrimaryAsset         common.Address
		SecondaryAsset       common.Address
		WithdrawTimeLock     *big.Int
		FastWithdrawDisabled bool
		MaxPositionAmount    *big.Int
		Insurance            common.Address
		FundingRateKeeper    common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.PrimaryAsset = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.SecondaryAsset = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.WithdrawTimeLock = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.FastWithdrawDisabled = *abi.ConvertType(out[3], new(bool)).(*bool)
	outstruct.MaxPositionAmount = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.Insurance = *abi.ConvertType(out[5], new(common.Address)).(*common.Address)
	outstruct.FundingRateKeeper = *abi.ConvertType(out[6], new(common.Address)).(*common.Address)

	return *outstruct, err

}

// State is a free data retrieval call binding the contract method 0xc19d93fb.
//
// Solidity: function state() view returns(address primaryAsset, address secondaryAsset, uint256 withdrawTimeLock, bool fastWithdrawDisabled, uint256 maxPositionAmount, address insurance, address fundingRateKeeper)
func (_Dealer *DealerSession) State() (struct {
	PrimaryAsset         common.Address
	SecondaryAsset       common.Address
	WithdrawTimeLock     *big.Int
	FastWithdrawDisabled bool
	MaxPositionAmount    *big.Int
	Insurance            common.Address
	FundingRateKeeper    common.Address
}, error) {
	return _Dealer.Contract.State(&_Dealer.CallOpts)
}

// State is a free data retrieval call binding the contract method 0xc19d93fb.
//
// Solidity: function state() view returns(address primaryAsset, address secondaryAsset, uint256 withdrawTimeLock, bool fastWithdrawDisabled, uint256 maxPositionAmount, address insurance, address fundingRateKeeper)
func (_Dealer *DealerCallerSession) State() (struct {
	PrimaryAsset         common.Address
	SecondaryAsset       common.Address
	WithdrawTimeLock     *big.Int
	FastWithdrawDisabled bool
	MaxPositionAmount    *big.Int
	Insurance            common.Address
	FundingRateKeeper    common.Address
}, error) {
	return _Dealer.Contract.State(&_Dealer.CallOpts)
}

// ApproveTrade is a paid mutator transaction binding the contract method 0x3cd9523d.
//
// Solidity: function approveTrade(address orderSender, bytes tradeData) returns(address[], int256[], int256[])
func (_Dealer *DealerTransactor) ApproveTrade(opts *bind.TransactOpts, orderSender common.Address, tradeData []byte) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "approveTrade", orderSender, tradeData)
}

// ApproveTrade is a paid mutator transaction binding the contract method 0x3cd9523d.
//
// Solidity: function approveTrade(address orderSender, bytes tradeData) returns(address[], int256[], int256[])
func (_Dealer *DealerSession) ApproveTrade(orderSender common.Address, tradeData []byte) (*types.Transaction, error) {
	return _Dealer.Contract.ApproveTrade(&_Dealer.TransactOpts, orderSender, tradeData)
}

// ApproveTrade is a paid mutator transaction binding the contract method 0x3cd9523d.
//
// Solidity: function approveTrade(address orderSender, bytes tradeData) returns(address[], int256[], int256[])
func (_Dealer *DealerTransactorSession) ApproveTrade(orderSender common.Address, tradeData []byte) (*types.Transaction, error) {
	return _Dealer.Contract.ApproveTrade(&_Dealer.TransactOpts, orderSender, tradeData)
}

// Deposit is a paid mutator transaction binding the contract method 0x8dbdbe6d.
//
// Solidity: function deposit(uint256 primaryAmount, uint256 secondaryAmount, address to) returns()
func (_Dealer *DealerTransactor) Deposit(opts *bind.TransactOpts, primaryAmount *big.Int, secondaryAmount *big.Int, to common.Address) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "deposit", primaryAmount, secondaryAmount, to)
}

// Deposit is a paid mutator transaction binding the contract method 0x8dbdbe6d.
//
// Solidity: function deposit(uint256 primaryAmount, uint256 secondaryAmount, address to) returns()
func (_Dealer *DealerSession) Deposit(primaryAmount *big.Int, secondaryAmount *big.Int, to common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.Deposit(&_Dealer.TransactOpts, primaryAmount, secondaryAmount, to)
}

// Deposit is a paid mutator transaction binding the contract method 0x8dbdbe6d.
//
// Solidity: function deposit(uint256 primaryAmount, uint256 secondaryAmount, address to) returns()
func (_Dealer *DealerTransactorSession) Deposit(primaryAmount *big.Int, secondaryAmount *big.Int, to common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.Deposit(&_Dealer.TransactOpts, primaryAmount, secondaryAmount, to)
}

// ExecuteWithdraw is a paid mutator transaction binding the contract method 0x43138789.
//
// Solidity: function executeWithdraw(address to) returns()
func (_Dealer *DealerTransactor) ExecuteWithdraw(opts *bind.TransactOpts, to common.Address) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "executeWithdraw", to)
}

// ExecuteWithdraw is a paid mutator transaction binding the contract method 0x43138789.
//
// Solidity: function executeWithdraw(address to) returns()
func (_Dealer *DealerSession) ExecuteWithdraw(to common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.ExecuteWithdraw(&_Dealer.TransactOpts, to)
}

// ExecuteWithdraw is a paid mutator transaction binding the contract method 0x43138789.
//
// Solidity: function executeWithdraw(address to) returns()
func (_Dealer *DealerTransactorSession) ExecuteWithdraw(to common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.ExecuteWithdraw(&_Dealer.TransactOpts, to)
}

// FastWithdraw is a paid mutator transaction binding the contract method 0xd84208cc.
//
// Solidity: function fastWithdraw(address to, uint256 primaryAmount, uint256 secondaryAmount) returns()
func (_Dealer *DealerTransactor) FastWithdraw(opts *bind.TransactOpts, to common.Address, primaryAmount *big.Int, secondaryAmount *big.Int) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "fastWithdraw", to, primaryAmount, secondaryAmount)
}

// FastWithdraw is a paid mutator transaction binding the contract method 0xd84208cc.
//
// Solidity: function fastWithdraw(address to, uint256 primaryAmount, uint256 secondaryAmount) returns()
func (_Dealer *DealerSession) FastWithdraw(to common.Address, primaryAmount *big.Int, secondaryAmount *big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.FastWithdraw(&_Dealer.TransactOpts, to, primaryAmount, secondaryAmount)
}

// FastWithdraw is a paid mutator transaction binding the contract method 0xd84208cc.
//
// Solidity: function fastWithdraw(address to, uint256 primaryAmount, uint256 secondaryAmount) returns()
func (_Dealer *DealerTransactorSession) FastWithdraw(to common.Address, primaryAmount *big.Int, secondaryAmount *big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.FastWithdraw(&_Dealer.TransactOpts, to, primaryAmount, secondaryAmount)
}

// HandleBadDebt is a paid mutator transaction binding the contract method 0x11fd52e1.
//
// Solidity: function handleBadDebt(address liquidatedTrader) returns()
func (_Dealer *DealerTransactor) HandleBadDebt(opts *bind.TransactOpts, liquidatedTrader common.Address) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "handleBadDebt", liquidatedTrader)
}

// HandleBadDebt is a paid mutator transaction binding the contract method 0x11fd52e1.
//
// Solidity: function handleBadDebt(address liquidatedTrader) returns()
func (_Dealer *DealerSession) HandleBadDebt(liquidatedTrader common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.HandleBadDebt(&_Dealer.TransactOpts, liquidatedTrader)
}

// HandleBadDebt is a paid mutator transaction binding the contract method 0x11fd52e1.
//
// Solidity: function handleBadDebt(address liquidatedTrader) returns()
func (_Dealer *DealerTransactorSession) HandleBadDebt(liquidatedTrader common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.HandleBadDebt(&_Dealer.TransactOpts, liquidatedTrader)
}

// OpenPosition is a paid mutator transaction binding the contract method 0xebbeba6e.
//
// Solidity: function openPosition(address trader) returns()
func (_Dealer *DealerTransactor) OpenPosition(opts *bind.TransactOpts, trader common.Address) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "openPosition", trader)
}

// OpenPosition is a paid mutator transaction binding the contract method 0xebbeba6e.
//
// Solidity: function openPosition(address trader) returns()
func (_Dealer *DealerSession) OpenPosition(trader common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.OpenPosition(&_Dealer.TransactOpts, trader)
}

// OpenPosition is a paid mutator transaction binding the contract method 0xebbeba6e.
//
// Solidity: function openPosition(address trader) returns()
func (_Dealer *DealerTransactorSession) OpenPosition(trader common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.OpenPosition(&_Dealer.TransactOpts, trader)
}

// RealizePnl is a paid mutator transaction binding the contract method 0x15e3ecb9.
//
// Solidity: function realizePnl(address trader, int256 pnl) returns()
func (_Dealer *DealerTransactor) RealizePnl(opts *bind.TransactOpts, trader common.Address, pnl *big.Int) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "realizePnl", trader, pnl)
}

// RealizePnl is a paid mutator transaction binding the contract method 0x15e3ecb9.
//
// Solidity: function realizePnl(address trader, int256 pnl) returns()
func (_Dealer *DealerSession) RealizePnl(trader common.Address, pnl *big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.RealizePnl(&_Dealer.TransactOpts, trader, pnl)
}

// RealizePnl is a paid mutator transaction binding the contract method 0x15e3ecb9.
//
// Solidity: function realizePnl(address trader, int256 pnl) returns()
func (_Dealer *DealerTransactorSession) RealizePnl(trader common.Address, pnl *big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.RealizePnl(&_Dealer.TransactOpts, trader, pnl)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Dealer *DealerTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Dealer *DealerSession) RenounceOwnership() (*types.Transaction, error) {
	return _Dealer.Contract.RenounceOwnership(&_Dealer.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Dealer *DealerTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Dealer.Contract.RenounceOwnership(&_Dealer.TransactOpts)
}

// RequestLiquidation is a paid mutator transaction binding the contract method 0xbe5b0371.
//
// Solidity: function requestLiquidation(address executor, address liquidator, address liquidatedTrader, int256 requestPaperAmount) returns(int256 liqtorPaperChange, int256 liqtorCreditChange, int256 liqedPaperChange, int256 liqedCreditChange)
func (_Dealer *DealerTransactor) RequestLiquidation(opts *bind.TransactOpts, executor common.Address, liquidator common.Address, liquidatedTrader common.Address, requestPaperAmount *big.Int) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "requestLiquidation", executor, liquidator, liquidatedTrader, requestPaperAmount)
}

// RequestLiquidation is a paid mutator transaction binding the contract method 0xbe5b0371.
//
// Solidity: function requestLiquidation(address executor, address liquidator, address liquidatedTrader, int256 requestPaperAmount) returns(int256 liqtorPaperChange, int256 liqtorCreditChange, int256 liqedPaperChange, int256 liqedCreditChange)
func (_Dealer *DealerSession) RequestLiquidation(executor common.Address, liquidator common.Address, liquidatedTrader common.Address, requestPaperAmount *big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.RequestLiquidation(&_Dealer.TransactOpts, executor, liquidator, liquidatedTrader, requestPaperAmount)
}

// RequestLiquidation is a paid mutator transaction binding the contract method 0xbe5b0371.
//
// Solidity: function requestLiquidation(address executor, address liquidator, address liquidatedTrader, int256 requestPaperAmount) returns(int256 liqtorPaperChange, int256 liqtorCreditChange, int256 liqedPaperChange, int256 liqedCreditChange)
func (_Dealer *DealerTransactorSession) RequestLiquidation(executor common.Address, liquidator common.Address, liquidatedTrader common.Address, requestPaperAmount *big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.RequestLiquidation(&_Dealer.TransactOpts, executor, liquidator, liquidatedTrader, requestPaperAmount)
}

// RequestWithdraw is a paid mutator transaction binding the contract method 0xe20cd4e3.
//
// Solidity: function requestWithdraw(uint256 primaryAmount, uint256 secondaryAmount) returns()
func (_Dealer *DealerTransactor) RequestWithdraw(opts *bind.TransactOpts, primaryAmount *big.Int, secondaryAmount *big.Int) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "requestWithdraw", primaryAmount, secondaryAmount)
}

// RequestWithdraw is a paid mutator transaction binding the contract method 0xe20cd4e3.
//
// Solidity: function requestWithdraw(uint256 primaryAmount, uint256 secondaryAmount) returns()
func (_Dealer *DealerSession) RequestWithdraw(primaryAmount *big.Int, secondaryAmount *big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.RequestWithdraw(&_Dealer.TransactOpts, primaryAmount, secondaryAmount)
}

// RequestWithdraw is a paid mutator transaction binding the contract method 0xe20cd4e3.
//
// Solidity: function requestWithdraw(uint256 primaryAmount, uint256 secondaryAmount) returns()
func (_Dealer *DealerTransactorSession) RequestWithdraw(primaryAmount *big.Int, secondaryAmount *big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.RequestWithdraw(&_Dealer.TransactOpts, primaryAmount, secondaryAmount)
}

// SetFundingRateKeeper is a paid mutator transaction binding the contract method 0x37a92fe4.
//
// Solidity: function setFundingRateKeeper(address newKeeper) returns()
func (_Dealer *DealerTransactor) SetFundingRateKeeper(opts *bind.TransactOpts, newKeeper common.Address) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "setFundingRateKeeper", newKeeper)
}

// SetFundingRateKeeper is a paid mutator transaction binding the contract method 0x37a92fe4.
//
// Solidity: function setFundingRateKeeper(address newKeeper) returns()
func (_Dealer *DealerSession) SetFundingRateKeeper(newKeeper common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.SetFundingRateKeeper(&_Dealer.TransactOpts, newKeeper)
}

// SetFundingRateKeeper is a paid mutator transaction binding the contract method 0x37a92fe4.
//
// Solidity: function setFundingRateKeeper(address newKeeper) returns()
func (_Dealer *DealerTransactorSession) SetFundingRateKeeper(newKeeper common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.SetFundingRateKeeper(&_Dealer.TransactOpts, newKeeper)
}

// SetInsurance is a paid mutator transaction binding the contract method 0x6cd7751f.
//
// Solidity: function setInsurance(address _insurance) returns()
func (_Dealer *DealerTransactor) SetInsurance(opts *bind.TransactOpts, _insurance common.Address) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "setInsurance", _insurance)
}

// SetInsurance is a paid mutator transaction binding the contract method 0x6cd7751f.
//
// Solidity: function setInsurance(address _insurance) returns()
func (_Dealer *DealerSession) SetInsurance(_insurance common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.SetInsurance(&_Dealer.TransactOpts, _insurance)
}

// SetInsurance is a paid mutator transaction binding the contract method 0x6cd7751f.
//
// Solidity: function setInsurance(address _insurance) returns()
func (_Dealer *DealerTransactorSession) SetInsurance(_insurance common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.SetInsurance(&_Dealer.TransactOpts, _insurance)
}

// SetMaxPositionAmount is a paid mutator transaction binding the contract method 0x90b74b44.
//
// Solidity: function setMaxPositionAmount(uint256 amount) returns()
func (_Dealer *DealerTransactor) SetMaxPositionAmount(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "setMaxPositionAmount", amount)
}

// SetMaxPositionAmount is a paid mutator transaction binding the contract method 0x90b74b44.
//
// Solidity: function setMaxPositionAmount(uint256 amount) returns()
func (_Dealer *DealerSession) SetMaxPositionAmount(amount *big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.SetMaxPositionAmount(&_Dealer.TransactOpts, amount)
}

// SetMaxPositionAmount is a paid mutator transaction binding the contract method 0x90b74b44.
//
// Solidity: function setMaxPositionAmount(uint256 amount) returns()
func (_Dealer *DealerTransactorSession) SetMaxPositionAmount(amount *big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.SetMaxPositionAmount(&_Dealer.TransactOpts, amount)
}

// SetOperator is a paid mutator transaction binding the contract method 0x558a7297.
//
// Solidity: function setOperator(address operator, bool isValid) returns()
func (_Dealer *DealerTransactor) SetOperator(opts *bind.TransactOpts, operator common.Address, isValid bool) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "setOperator", operator, isValid)
}

// SetOperator is a paid mutator transaction binding the contract method 0x558a7297.
//
// Solidity: function setOperator(address operator, bool isValid) returns()
func (_Dealer *DealerSession) SetOperator(operator common.Address, isValid bool) (*types.Transaction, error) {
	return _Dealer.Contract.SetOperator(&_Dealer.TransactOpts, operator, isValid)
}

// SetOperator is a paid mutator transaction binding the contract method 0x558a7297.
//
// Solidity: function setOperator(address operator, bool isValid) returns()
func (_Dealer *DealerTransactorSession) SetOperator(operator common.Address, isValid bool) (*types.Transaction, error) {
	return _Dealer.Contract.SetOperator(&_Dealer.TransactOpts, operator, isValid)
}

// SetOrderSender is a paid mutator transaction binding the contract method 0x85965903.
//
// Solidity: function setOrderSender(address orderSender, bool isValid) returns()
func (_Dealer *DealerTransactor) SetOrderSender(opts *bind.TransactOpts, orderSender common.Address, isValid bool) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "setOrderSender", orderSender, isValid)
}

// SetOrderSender is a paid mutator transaction binding the contract method 0x85965903.
//
// Solidity: function setOrderSender(address orderSender, bool isValid) returns()
func (_Dealer *DealerSession) SetOrderSender(orderSender common.Address, isValid bool) (*types.Transaction, error) {
	return _Dealer.Contract.SetOrderSender(&_Dealer.TransactOpts, orderSender, isValid)
}

// SetOrderSender is a paid mutator transaction binding the contract method 0x85965903.
//
// Solidity: function setOrderSender(address orderSender, bool isValid) returns()
func (_Dealer *DealerTransactorSession) SetOrderSender(orderSender common.Address, isValid bool) (*types.Transaction, error) {
	return _Dealer.Contract.SetOrderSender(&_Dealer.TransactOpts, orderSender, isValid)
}

// SetPerpRiskParams is a paid mutator transaction binding the contract method 0xc3d155bd.
//
// Solidity: function setPerpRiskParams(address perp, (uint256,uint256,uint256,uint256,address,string,bool) params) returns()
func (_Dealer *DealerTransactor) SetPerpRiskParams(opts *bind.TransactOpts, perp common.Address, params TypesRiskParams) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "setPerpRiskParams", perp, params)
}

// SetPerpRiskParams is a paid mutator transaction binding the contract method 0xc3d155bd.
//
// Solidity: function setPerpRiskParams(address perp, (uint256,uint256,uint256,uint256,address,string,bool) params) returns()
func (_Dealer *DealerSession) SetPerpRiskParams(perp common.Address, params TypesRiskParams) (*types.Transaction, error) {
	return _Dealer.Contract.SetPerpRiskParams(&_Dealer.TransactOpts, perp, params)
}

// SetPerpRiskParams is a paid mutator transaction binding the contract method 0xc3d155bd.
//
// Solidity: function setPerpRiskParams(address perp, (uint256,uint256,uint256,uint256,address,string,bool) params) returns()
func (_Dealer *DealerTransactorSession) SetPerpRiskParams(perp common.Address, params TypesRiskParams) (*types.Transaction, error) {
	return _Dealer.Contract.SetPerpRiskParams(&_Dealer.TransactOpts, perp, params)
}

// SetWithdrawTimeLock is a paid mutator transaction binding the contract method 0x1ac69fce.
//
// Solidity: function setWithdrawTimeLock(uint256 timeLock) returns()
func (_Dealer *DealerTransactor) SetWithdrawTimeLock(opts *bind.TransactOpts, timeLock *big.Int) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "setWithdrawTimeLock", timeLock)
}

// SetWithdrawTimeLock is a paid mutator transaction binding the contract method 0x1ac69fce.
//
// Solidity: function setWithdrawTimeLock(uint256 timeLock) returns()
func (_Dealer *DealerSession) SetWithdrawTimeLock(timeLock *big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.SetWithdrawTimeLock(&_Dealer.TransactOpts, timeLock)
}

// SetWithdrawTimeLock is a paid mutator transaction binding the contract method 0x1ac69fce.
//
// Solidity: function setWithdrawTimeLock(uint256 timeLock) returns()
func (_Dealer *DealerTransactorSession) SetWithdrawTimeLock(timeLock *big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.SetWithdrawTimeLock(&_Dealer.TransactOpts, timeLock)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Dealer *DealerTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Dealer *DealerSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.TransferOwnership(&_Dealer.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Dealer *DealerTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Dealer.Contract.TransferOwnership(&_Dealer.TransactOpts, newOwner)
}

// UpdateFundingRate is a paid mutator transaction binding the contract method 0x854a8041.
//
// Solidity: function updateFundingRate(address[] perpList, int256[] rateList) returns()
func (_Dealer *DealerTransactor) UpdateFundingRate(opts *bind.TransactOpts, perpList []common.Address, rateList []*big.Int) (*types.Transaction, error) {
	return _Dealer.contract.Transact(opts, "updateFundingRate", perpList, rateList)
}

// UpdateFundingRate is a paid mutator transaction binding the contract method 0x854a8041.
//
// Solidity: function updateFundingRate(address[] perpList, int256[] rateList) returns()
func (_Dealer *DealerSession) UpdateFundingRate(perpList []common.Address, rateList []*big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.UpdateFundingRate(&_Dealer.TransactOpts, perpList, rateList)
}

// UpdateFundingRate is a paid mutator transaction binding the contract method 0x854a8041.
//
// Solidity: function updateFundingRate(address[] perpList, int256[] rateList) returns()
func (_Dealer *DealerTransactorSession) UpdateFundingRate(perpList []common.Address, rateList []*big.Int) (*types.Transaction, error) {
	return _Dealer.Contract.UpdateFundingRate(&_Dealer.TransactOpts, perpList, rateList)
}

// DealerOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Dealer contract.
type DealerOwnershipTransferredIterator struct {
	Event *DealerOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DealerOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DealerOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DealerOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DealerOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DealerOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DealerOwnershipTransferred represents a OwnershipTransferred event raised by the Dealer contract.
type DealerOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Dealer *DealerFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*DealerOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Dealer.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &DealerOwnershipTransferredIterator{contract: _Dealer.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Dealer *DealerFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *DealerOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Dealer.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DealerOwnershipTransferred)
				if err := _Dealer.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Dealer *DealerFilterer) ParseOwnershipTransferred(log types.Log) (*DealerOwnershipTransferred, error) {
	event := new(DealerOwnershipTransferred)
	if err := _Dealer.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
