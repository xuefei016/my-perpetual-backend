// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package pricesource

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

// PriceSourceMetaData contains all meta data concerning the PriceSource contract.
var PriceSourceMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"getMarkPrice\",\"inputs\":[],\"outputs\":[{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"}]",
}

// PriceSourceABI is the input ABI used to generate the binding from.
// Deprecated: Use PriceSourceMetaData.ABI instead.
var PriceSourceABI = PriceSourceMetaData.ABI

// PriceSource is an auto generated Go binding around an Ethereum contract.
type PriceSource struct {
	PriceSourceCaller     // Read-only binding to the contract
	PriceSourceTransactor // Write-only binding to the contract
	PriceSourceFilterer   // Log filterer for contract events
}

// PriceSourceCaller is an auto generated read-only Go binding around an Ethereum contract.
type PriceSourceCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PriceSourceTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PriceSourceTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PriceSourceFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PriceSourceFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PriceSourceSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PriceSourceSession struct {
	Contract     *PriceSource      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PriceSourceCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PriceSourceCallerSession struct {
	Contract *PriceSourceCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// PriceSourceTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PriceSourceTransactorSession struct {
	Contract     *PriceSourceTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// PriceSourceRaw is an auto generated low-level Go binding around an Ethereum contract.
type PriceSourceRaw struct {
	Contract *PriceSource // Generic contract binding to access the raw methods on
}

// PriceSourceCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PriceSourceCallerRaw struct {
	Contract *PriceSourceCaller // Generic read-only contract binding to access the raw methods on
}

// PriceSourceTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PriceSourceTransactorRaw struct {
	Contract *PriceSourceTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPriceSource creates a new instance of PriceSource, bound to a specific deployed contract.
func NewPriceSource(address common.Address, backend bind.ContractBackend) (*PriceSource, error) {
	contract, err := bindPriceSource(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PriceSource{PriceSourceCaller: PriceSourceCaller{contract: contract}, PriceSourceTransactor: PriceSourceTransactor{contract: contract}, PriceSourceFilterer: PriceSourceFilterer{contract: contract}}, nil
}

// NewPriceSourceCaller creates a new read-only instance of PriceSource, bound to a specific deployed contract.
func NewPriceSourceCaller(address common.Address, caller bind.ContractCaller) (*PriceSourceCaller, error) {
	contract, err := bindPriceSource(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PriceSourceCaller{contract: contract}, nil
}

// NewPriceSourceTransactor creates a new write-only instance of PriceSource, bound to a specific deployed contract.
func NewPriceSourceTransactor(address common.Address, transactor bind.ContractTransactor) (*PriceSourceTransactor, error) {
	contract, err := bindPriceSource(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PriceSourceTransactor{contract: contract}, nil
}

// NewPriceSourceFilterer creates a new log filterer instance of PriceSource, bound to a specific deployed contract.
func NewPriceSourceFilterer(address common.Address, filterer bind.ContractFilterer) (*PriceSourceFilterer, error) {
	contract, err := bindPriceSource(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PriceSourceFilterer{contract: contract}, nil
}

// bindPriceSource binds a generic wrapper to an already deployed contract.
func bindPriceSource(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PriceSourceMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PriceSource *PriceSourceRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PriceSource.Contract.PriceSourceCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PriceSource *PriceSourceRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PriceSource.Contract.PriceSourceTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PriceSource *PriceSourceRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PriceSource.Contract.PriceSourceTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PriceSource *PriceSourceCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PriceSource.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PriceSource *PriceSourceTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PriceSource.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PriceSource *PriceSourceTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PriceSource.Contract.contract.Transact(opts, method, params...)
}

// GetMarkPrice is a free data retrieval call binding the contract method 0x60409632.
//
// Solidity: function getMarkPrice() view returns(uint256 price)
func (_PriceSource *PriceSourceCaller) GetMarkPrice(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PriceSource.contract.Call(opts, &out, "getMarkPrice")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMarkPrice is a free data retrieval call binding the contract method 0x60409632.
//
// Solidity: function getMarkPrice() view returns(uint256 price)
func (_PriceSource *PriceSourceSession) GetMarkPrice() (*big.Int, error) {
	return _PriceSource.Contract.GetMarkPrice(&_PriceSource.CallOpts)
}

// GetMarkPrice is a free data retrieval call binding the contract method 0x60409632.
//
// Solidity: function getMarkPrice() view returns(uint256 price)
func (_PriceSource *PriceSourceCallerSession) GetMarkPrice() (*big.Int, error) {
	return _PriceSource.Contract.GetMarkPrice(&_PriceSource.CallOpts)
}
