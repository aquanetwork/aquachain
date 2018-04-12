// Copyright 2016 The aquachain Authors
// This file is part of the aquachain library.
//
// The aquachain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The aquachain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the aquachain library. If not, see <http://www.gnu.org/licenses/>.

package params

import (
	"fmt"
	"math/big"

	"github.com/aquanetwork/aquachain/common"
)

var (
	MainnetGenesisHash = common.HexToHash("0x381c8d2c3e3bc702533ee504d7621d510339cafd830028337a4b532ff27cd505") // Mainnet genesis hash to enforce below configs on
	TestnetGenesisHash = common.HexToHash("0x3116d55259a6babc9ecac1eee645cbbdeaa4ee7aa7cab75693e21d2f0cf724a4") // Testnet genesis hash to enforce below configs on
)
var (
	AquachainHF = ForkMap{
		0: big.NewInt(3000),  //  hf0 had no changes
		1: big.NewInt(3600),  // increase min difficulty to the next multiple of 2048
		2: big.NewInt(7200),  // use simple difficulty algo (240 seconds)
		3: big.NewInt(13026), // increase min difficulty for anticipation of gpu mining
		4: big.NewInt(21800), // fix bad genesis alloc and improve p2p
		5: big.NewInt(22800), // move to keccak256(argon2id())
	}
	TestnetHF = ForkMap{
		0: big.NewInt(0),  //  hf0 had no changes
		1: big.NewInt(-1), // increase min difficulty to the next multiple of 2048
		2: big.NewInt(0),  // use simple difficulty algo (240 seconds)
		3: big.NewInt(-1), // increase min difficulty for anticipation of gpu mining
		4: big.NewInt(1),  //
		5: big.NewInt(1),
	}
)
var (
	// MainnetChainConfig is the chain parameters to run a node on the main network.
	MainnetChainConfig = &ChainConfig{
		ChainId:         big.NewInt(61717561),
		HomesteadBlock:  big.NewInt(0),
		EIP150Block:     big.NewInt(0),
		Aquahash:        new(AquahashConfig),
		ConsensusConfig: new(ConsensusConfig),
		HF:              AquachainHF,
	}

	// TestnetChainConfig contains the chain parameters to run a node on the Ropsten test network.
	TestnetChainConfig = &ChainConfig{
		ChainId:         big.NewInt(617175610),
		HomesteadBlock:  big.NewInt(0),
		EIP150Block:     big.NewInt(0),
		Aquahash:        new(AquahashConfig),
		ConsensusConfig: new(ConsensusConfig),
		HF:              TestnetHF,
	}

	// Testnet2ChainConfig contains the chain parameters to run a node on the Rinkeby test network.
	Testnet2ChainConfig = &ChainConfig{
		ChainId:         big.NewInt(617175611),
		HomesteadBlock:  big.NewInt(0),
		EIP150Block:     big.NewInt(0),
		Aquahash:        new(AquahashConfig),
		ConsensusConfig: new(ConsensusConfig),
		HF:              TestnetHF,
	}

	// AllAquahashProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the AquaChain core developers into the Aquahash consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllAquahashProtocolChanges = &ChainConfig{big.NewInt(1337), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, new(AquahashConfig), nil, new(ConsensusConfig), AquachainHF}

	// AllCliqueProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the AquaChain core developers into the Clique consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllCliqueProtocolChanges = &ChainConfig{big.NewInt(1337), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, nil, &CliqueConfig{Period: 0, Epoch: 30000}, new(ConsensusConfig), AquachainHF}

	TestChainConfig = &ChainConfig{big.NewInt(1), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, new(AquahashConfig), nil, new(ConsensusConfig), TestnetHF}
	TestRules       = TestChainConfig.Rules(new(big.Int))
)

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainId *big.Int `json:"chainId"` // Chain id identifies the current chain and is used for replay protection

	HomesteadBlock *big.Int `json:"homesteadBlock,omitempty"` // Homestead switch block (nil = no fork, 0 = already homestead)

	// HF Scheduled Maintenance Hardforks
	//HF map[int]*big.Int `json:"hfmap,omitempty"` // map of HF block numbers

	// et junk to remove
	DAOForkBlock   *big.Int `json:"daoForkBlock,omitempty"`   // TheDAO hard-fork switch block (nil = no fork)
	DAOForkSupport bool     `json:"daoForkSupport,omitempty"` // Whether the nodes supports or opposes the DAO hard-fork

	// EIP150 implements the Gas price changes (https://github.com/aquanetwork/EIPs/issues/150)
	EIP150Block *big.Int    `json:"eip150Block,omitempty"` // EIP150 HF block (nil = no fork)
	EIP150Hash  common.Hash `json:"eip150Hash,omitempty"`  // EIP150 HF hash (needed for header only clients as only gas pricing changed)

	EIP155Block *big.Int `json:"eip155Block,omitempty"` // EIP155 HF block
	EIP158Block *big.Int `json:"eip158Block,omitempty"` // EIP158 HF block

	ByzantiumBlock      *big.Int `json:"byzantiumBlock,omitempty"`      // Byzantium switch block (nil = no fork, 0 = already on byzantium)
	ConstantinopleBlock *big.Int `json:"constantinopleBlock,omitempty"` // Constantinople switch block (nil = no fork, 0 = already activated)

	// Various consensus engines
	Aquahash        *AquahashConfig  `json:"aquahash,omitempty"`
	Clique          *CliqueConfig    `json:"clique,omitempty"`
	ConsensusConfig *ConsensusConfig `json:"consensus,omitempty"`
	HF              ForkMap          `json:"hf,omitempty"`
}

// AquahashConfig is the consensus engine configs for proof-of-work based sealing.
type AquahashConfig struct{}

// String implements the stringer interface, returning the consensus engine details.
func (c *AquahashConfig) String() string {
	return "aquahash"
}

// ConsensusConfig is the consensus engine configs for proof-of-work based sealing.
type ConsensusConfig struct {
}

// String implements the stringer interface, returning the consensus engine details.
func (c *ConsensusConfig) String() string {
	return "argonated"
}

// CliqueConfig is the consensus engine configs for proof-of-authority based sealing.
type CliqueConfig struct {
	Period uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch  uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
}

// String implements the stringer interface, returning the consensus engine details.
func (c *CliqueConfig) String() string {
	return "clique"
}

// String implements the fmt.Stringer interface.
func (c *ChainConfig) String() string {
	var engine interface{}
	switch {
	case c.ConsensusConfig != nil:
		engine = c.ConsensusConfig
	case c.Aquahash != nil:
		engine = c.Aquahash
	case c.Clique != nil:
		engine = c.Clique
	default:
		engine = "unknown"
	}
	return fmt.Sprintf("{ChainID: %v, Engine: %v, HF-Ready: %s}",
		c.ChainId,
		engine,
		c.HF,
	)
	// return fmt.Sprintf("{ChainID: %v Homestead: %v DAO: %v DAOSupport: %v EIP150: %v EIP155: %v EIP158: %v Byzantium: %v Constantinople: %v Engine: %v}",
	// 	c.ChainId,
	// 	c.HomesteadBlock,
	// 	c.DAOForkBlock,
	// 	c.DAOForkSupport,
	// 	c.EIP150Block,
	// 	c.EIP155Block,
	// 	c.EIP158Block,
	// 	c.ByzantiumBlock,
	// 	c.ConstantinopleBlock,
	// 	engine,
	// )
}

// IsHF returns whether num is either equal to the hf block or greater.
func (c *ChainConfig) IsHF(hf int, num *big.Int) bool {
	forkheight := c.HF[hf]
	if forkheight == nil || forkheight.Sign() == -1 {
		return false
	}
	return isForked(forkheight, num)
}

// GetHF returns the height of the requested hf (CAN BE NIL)
func (c *ChainConfig) GetHF(hf int) *big.Int {
	return c.HF[hf]
}

// NextHF returns the first next scheduled hard fork block number
func (c *ChainConfig) NextHF(cur *big.Int) *big.Int {
	if cur != nil {
		for _, height := range c.HF {
			if cur.Cmp(height) < 0 {
				return height
			}
		}
	}
	return nil
}

// IsHomestead returns whether num is either equal to the homestead block or greater.
func (c *ChainConfig) IsHomestead(num *big.Int) bool {
	return isForked(c.HomesteadBlock, num)
}

// IsDAO returns whether num is either equal to the DAO fork block or greater.
func (c *ChainConfig) IsDAOFork(num *big.Int) bool {
	return isForked(c.DAOForkBlock, num)
}

func (c *ChainConfig) IsEIP150(num *big.Int) bool {
	return isForked(c.EIP150Block, num)
}

func (c *ChainConfig) IsEIP155(num *big.Int) bool {
	return isForked(c.EIP155Block, num)
}

func (c *ChainConfig) IsEIP158(num *big.Int) bool {
	return isForked(c.EIP158Block, num)
}

func (c *ChainConfig) IsByzantium(num *big.Int) bool {
	return isForked(c.ByzantiumBlock, num)
}

func (c *ChainConfig) IsConstantinople(num *big.Int) bool {
	return isForked(c.ConstantinopleBlock, num)
}

// GasTable returns the gas table corresponding to the current phase.
//
// The returned GasTable's fields shouldn't, under any circumstances, be changed.
func (c *ChainConfig) GasTable(num *big.Int) GasTable {
	if num == nil {
		return GasTableHomestead
	}
	switch {
	case c.IsHF(1, num):
		return GasTableHF1
	default:
		return GasTableHomestead
	}
}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *ChainConfig) CheckCompatible(newcfg *ChainConfig, height uint64) *ConfigCompatError {
	bhead := new(big.Int).SetUint64(height)

	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead)
		if err == nil || (lasterr != nil && err.RewindTo == lasterr.RewindTo) {
			break
		}
		lasterr = err
		bhead.SetUint64(err.RewindTo)
	}
	return lasterr
}

func (c *ChainConfig) checkCompatible(newcfg *ChainConfig, head *big.Int) *ConfigCompatError {
	if isForkIncompatible(c.HomesteadBlock, newcfg.HomesteadBlock, head) {
		return newCompatError("Homestead fork block", c.HomesteadBlock, newcfg.HomesteadBlock)
	}
	if isForkIncompatible(c.DAOForkBlock, newcfg.DAOForkBlock, head) {
		return newCompatError("DAO fork block", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if c.IsDAOFork(head) && c.DAOForkSupport != newcfg.DAOForkSupport {
		return newCompatError("DAO fork support flag", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if isForkIncompatible(c.EIP150Block, newcfg.EIP150Block, head) {
		return newCompatError("EIP150 fork block", c.EIP150Block, newcfg.EIP150Block)
	}
	if isForkIncompatible(c.EIP155Block, newcfg.EIP155Block, head) {
		return newCompatError("EIP155 fork block", c.EIP155Block, newcfg.EIP155Block)
	}
	if isForkIncompatible(c.EIP158Block, newcfg.EIP158Block, head) {
		return newCompatError("EIP158 fork block", c.EIP158Block, newcfg.EIP158Block)
	}
	if c.IsEIP158(head) && !configNumEqual(c.ChainId, newcfg.ChainId) {
		return newCompatError("EIP158 chain ID", c.EIP158Block, newcfg.EIP158Block)
	}
	if isForkIncompatible(c.ByzantiumBlock, newcfg.ByzantiumBlock, head) {
		return newCompatError("Byzantium fork block", c.ByzantiumBlock, newcfg.ByzantiumBlock)
	}
	if isForkIncompatible(c.ConstantinopleBlock, newcfg.ConstantinopleBlock, head) {
		return newCompatError("Constantinople fork block", c.ConstantinopleBlock, newcfg.ConstantinopleBlock)
	}
	return nil
}

// isForkIncompatible returns true if a fork scheduled at s1 cannot be rescheduled to
// block s2 because head is already past the fork.
func isForkIncompatible(s1, s2, head *big.Int) bool {
	return (isForked(s1, head) || isForked(s2, head)) && !configNumEqual(s1, s2)
}

// isForked returns whether a fork scheduled at block s is active at the given head block.
func isForked(s, head *big.Int) bool {
	if s == nil || head == nil {
		return false
	}
	return s.Cmp(head) <= 0
}

func configNumEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string
	// block numbers of the stored and new configurations
	StoredConfig, NewConfig *big.Int
	// the block number to which the local chain must be rewound to correct the error
	RewindTo uint64
}

func newCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{what, storedblock, newblock, 0}
	if rew != nil && rew.Sign() > 0 {
		err.RewindTo = rew.Uint64() - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	return fmt.Sprintf("mismatching %s in database (have %d, want %d, rewindto %d)", err.What, err.StoredConfig, err.NewConfig, err.RewindTo)
}

// Rules wraps ChainConfig and is merely syntatic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	ChainId                                   *big.Int
	IsHomestead, IsEIP150, IsEIP155, IsEIP158 bool
	IsByzantium                               bool
}

func (c *ChainConfig) Rules(num *big.Int) Rules {
	chainId := c.ChainId
	if chainId == nil {
		chainId = new(big.Int)
	}
	return Rules{ChainId: new(big.Int).Set(chainId), IsHomestead: c.IsHomestead(num), IsEIP150: c.IsEIP150(num), IsEIP155: c.IsEIP155(num), IsEIP158: c.IsEIP158(num), IsByzantium: c.IsByzantium(num)}
}
